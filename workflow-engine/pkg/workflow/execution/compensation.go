package execution

import (
	"context"
	"log"
	"sort"

	"github.com/google/uuid"

	executionModel "github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
)

func (e *Engine) compensate(
	ctx context.Context,
	execID uuid.UUID,
	nodes []executionModel.ExecutionNode,
) error {

	log.Printf("[Compensation] Starting compensation: execution_id=%s", execID)

	// Record high-level compensation start event
	_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
		ExecutionID: execID,
		EventType:   "COMPENSATION_STARTED",
		Message:     "Compensation started",
		Payload: map[string]any{
			"note": "best-effort compensation triggered after failure",
		},
	})

	// Reload workflow + DAG (safe & explicit)
	exec, err := e.ExecStore.Get(ctx, "", execID)
	if err != nil {
		return err
	}

	wf, err := e.WorkflowReg.Get(ctx, exec.ProjectID, exec.WorkflowID)
	if err != nil {
		return err
	}

	graph := dag.Build(wf.Def)

	//Collect succeeded nodes
	completed := []executionModel.ExecutionNode{}
	for _, n := range nodes {
		if n.Status == executionModel.NodeSucceeded {
			completed = append(completed, n)
		}
	}

	//Reverse order (LIFO by completion time). Handle possible nil timestamps safely.
	sort.Slice(completed, func(i, j int) bool {
		ti := completed[i].CompletedAt
		tj := completed[j].CompletedAt

		// If both are nil, keep original order
		if ti == nil && tj == nil {
			return false
		}
		// Treat nil completion time as "older" than any non-nil time
		if ti == nil {
			return false
		}
		if tj == nil {
			return true
		}

		return ti.After(*tj)
	})

	// Execute compensation
	for _, node := range completed {
		dagNode := graph.Nodes[dag.NodeID(node.NodeID)]
		if dagNode == nil || dagNode.Compensate == nil {
			continue
		}

		log.Printf(
			"[Compensation] Executing compensation: execution_id=%s, node_id=%s, uses=%s",
			execID,
			node.NodeID,
			dagNode.Compensate.Uses,
		)

		// Record per-node compensation start event
		_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
			ExecutionID: execID,
			NodeID:      &node.NodeID,
			EventType:   "COMPENSATION_NODE_STARTED",
			Message:     "Compensation step started",
			Payload: map[string]any{
				"compensate_uses": dagNode.Compensate.Uses,
			},
		})

		mod, err := e.ModuleReg.Resolve(ctx, "", dagNode.Compensate.Uses)
		if err != nil {
			log.Printf("[Compensation] Module resolve failed: %v", err)
			_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
				ExecutionID: execID,
				NodeID:      &node.NodeID,
				EventType:   "COMPENSATION_NODE_FAILED",
				Message:     "Compensation module resolve failed",
				Payload: map[string]any{
					"compensate_uses": dagNode.Compensate.Uses,
					"error":           err.Error(),
				},
			})
			continue
		}

		execImpl, err := executor.Get(mod.Runtime)
		if err != nil {
			log.Printf("[Compensation] Executor load failed: %v", err)
			_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
				ExecutionID: execID,
				NodeID:      &node.NodeID,
				EventType:   "COMPENSATION_NODE_FAILED",
				Message:     "Compensation executor load failed",
				Payload: map[string]any{
					"compensate_uses": dagNode.Compensate.Uses,
					"error":           err.Error(),
				},
			})
			continue
		}

		_, err = execImpl.Execute(
			ctx,
			&dag.Node{
				Uses: dagNode.Compensate.Uses,
				With: dagNode.Compensate.With,
			},
			node.Output,
		)

		if err != nil {
			log.Printf(
				"[Compensation] Failed: execution_id=%s, node_id=%s, error=%v",
				execID,
				node.NodeID,
				err,
			)
			_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
				ExecutionID: execID,
				NodeID:      &node.NodeID,
				EventType:   "COMPENSATION_NODE_FAILED",
				Message:     "Compensation step execution failed",
				Payload: map[string]any{
				 "compensate_uses": dagNode.Compensate.Uses,
				 "error":           err.Error(),
				},
			})
			continue
		}

		// Record per-node compensation success
		_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
			ExecutionID: execID,
			NodeID:      &node.NodeID,
			EventType:   "COMPENSATION_NODE_SUCCEEDED",
			Message:     "Compensation step succeeded",
			Payload: map[string]any{
				"compensate_uses": dagNode.Compensate.Uses,
			},
		})
	}

	log.Printf("[Compensation] Completed: execution_id=%s", execID)
	// Record high-level completion event
	_ = e.EventStore.Append(ctx, &executionModel.ExecutionEvent{
		ExecutionID: execID,
		EventType:   "COMPENSATION_COMPLETED",
		Message:     "Compensation completed",
	})
	return nil
}
