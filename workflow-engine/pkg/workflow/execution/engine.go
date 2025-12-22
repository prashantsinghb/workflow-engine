package execution

import (
	"context"
	"time"

	"github.com/google/uuid"

	execution "github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	wfRegistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

type Engine struct {
	ExecStore   execution.ExecutionStore
	NodeStore   execution.NodeStore
	EventStore  execution.EventStore
	WorkflowReg wfRegistry.WorkflowStore
	ModuleReg   *registry.ModuleRegistry
}

const RetryDelay = 1 * time.Second

// StartExecution starts a workflow execution
func (e *Engine) StartExecution(ctx context.Context, projectID string, execID uuid.UUID) error {
	// Load execution
	exec, err := e.ExecStore.Get(ctx, projectID, execID)
	if err != nil {
		return err
	}

	// Mark as running
	if err := e.ExecStore.MarkRunning(ctx, execID); err != nil {
		return err
	}

	// Load workflow
	wf, err := e.WorkflowReg.Get(ctx, exec.ProjectID, exec.WorkflowID)
	if err != nil {
		return err
	}

	graph := dag.Build(wf.Def)

	// Ensure nodes exist in DB
	if err := e.ensureNodes(ctx, execID, graph); err != nil {
		return err
	}

	// Reconcile DAG
	return e.reconcile(ctx, execID, graph)
}

// ensureNodes ensures all DAG nodes exist in DB
func (e *Engine) ensureNodes(ctx context.Context, execID uuid.UUID, graph *dag.Graph) error {
	for id, node := range graph.Nodes {
		n := &execution.ExecutionNode{
			ExecutionID:  execID,
			NodeID:       string(id),
			ExecutorType: node.Uses,
			Status:       execution.NodePending,
			MaxAttempts:  3, // default retries if needed
			Attempt:      0,
		}
		if err := e.NodeStore.Upsert(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

// reconcile loops through DAG and executes ready nodes
func (e *Engine) reconcile(ctx context.Context, execID uuid.UUID, graph *dag.Graph) error {
	for {
		// Load all nodes for execution
		nodes, err := e.NodeStore.ListByExecution(ctx, execID)
		if err != nil {
			return err
		}

		// Track progress
		progress := false

		for i := range nodes {
			node := &nodes[i]
			if node.Status != execution.NodePending && node.Status != execution.NodeRetrying {
				continue
			}

			// Check dependencies
			if !e.dependenciesDone(node, nodes, graph) {
				continue
			}

			progress = true
			if err := e.executeNode(ctx, node, graph); err != nil {
				// Node marked failed inside executeNode
				continue
			}
		}

		// Check if DAG is completed
		if e.isCompleted(nodes, execID) {
			return nil
		}

		if !progress {
			// Deadlock detected
			return e.ExecStore.MarkFailed(ctx, execID, map[string]any{"reason": "deadlock detected in DAG execution"})
		}

		// Sleep briefly to avoid busy loop
		time.Sleep(RetryDelay)
	}
}

// dependenciesDone returns true if all dependencies of a node are completed successfully
func (e *Engine) dependenciesDone(node *execution.ExecutionNode, allNodes []execution.ExecutionNode, graph *dag.Graph) bool {

	graphNode, ok := graph.Nodes[dag.NodeID(node.NodeID)]
	if !ok {
		return false // Node not found in graph
	}

	// Create a map of completed nodes for quick lookup
	completed := make(map[string]bool)
	for _, n := range allNodes {
		if n.Status == execution.NodeSucceeded {
			completed[n.NodeID] = true
		}
	}

	// Check if all dependencies are completed
	for _, dep := range graphNode.Depends {
		if !completed[string(dep)] {
			return false
		}
	}

	return true
}

// executeNode executes a single node
func (e *Engine) executeNode(ctx context.Context, node *execution.ExecutionNode, graph *dag.Graph) error {
	node.Attempt++
	if err := e.NodeStore.MarkRunning(ctx, node.ExecutionID, node.NodeID); err != nil {
		return err
	}

	mod, err := e.ModuleReg.Resolve(ctx, "", node.ExecutorType)
	if err != nil {
		return e.failOrRetry(ctx, node, err)
	}

	execImpl, err := executor.Get(mod.Runtime)
	if err != nil {
		return e.failOrRetry(ctx, node, err)
	}

	out, err := execImpl.Execute(ctx, graph.Nodes[dag.NodeID(node.NodeID)], node.Input)
	if err != nil {
		return e.failOrRetry(ctx, node, err)
	}

	return e.NodeStore.MarkSucceeded(ctx, node.ExecutionID, node.NodeID, map[string]any{"output": out})
}

// failOrRetry handles retry logic or marks node failed
func (e *Engine) failOrRetry(ctx context.Context, node *execution.ExecutionNode, execErr error) error {
	if node.Attempt < node.MaxAttempts {
		node.Status = execution.NodeRetrying
		if err := e.NodeStore.Upsert(ctx, node); err != nil {
			return err
		}
		return nil
	}

	return e.NodeStore.MarkFailed(ctx, node.ExecutionID, node.NodeID, map[string]any{"message": execErr.Error()})
}

// isCompleted checks if all nodes have finished
func (e *Engine) isCompleted(nodes []execution.ExecutionNode, execID uuid.UUID) bool {
	allSucceeded := true
	for _, node := range nodes {
		switch node.Status {
		case execution.NodeFailed:
			_ = e.ExecStore.MarkFailed(context.Background(), execID, map[string]any{"reason": "one or more nodes failed"})
			return true
		case execution.NodePending, execution.NodeRetrying:
			allSucceeded = false
		}
	}

	if allSucceeded {
		// Collect outputs
		outputs := map[string]any{}
		for _, node := range nodes {
			outputs[node.NodeID] = node.Output
		}
		_ = e.ExecStore.MarkCompleted(context.Background(), execID, map[string]any{"outputs": outputs})
		return true
	}

	return false
}
