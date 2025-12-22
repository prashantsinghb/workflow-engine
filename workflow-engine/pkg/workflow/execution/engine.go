package execution

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("[Engine] Starting execution: execution_id=%s, project_id=%s", execID, projectID)

	// Load execution
	exec, err := e.ExecStore.Get(ctx, projectID, execID)
	if err != nil {
		log.Printf("[Engine] Failed to load execution: execution_id=%s, error=%v", execID, err)
		return err
	}

	// Mark as running
	if err := e.ExecStore.MarkRunning(ctx, execID); err != nil {
		log.Printf("[Engine] Failed to mark execution as running: execution_id=%s, error=%v", execID, err)
		return err
	}
	log.Printf("[Engine] Execution marked as running: execution_id=%s", execID)

	// Load workflow
	wf, err := e.WorkflowReg.Get(ctx, exec.ProjectID, exec.WorkflowID)
	if err != nil {
		log.Printf("[Engine] Failed to load workflow: execution_id=%s, workflow_id=%s, error=%v", execID, exec.WorkflowID, err)
		return err
	}
	log.Printf("[Engine] Workflow loaded: execution_id=%s, workflow_id=%s", execID, exec.WorkflowID)

	graph := dag.Build(wf.Def)
	log.Printf("[Engine] DAG built: execution_id=%s, node_count=%d", execID, len(graph.Nodes))

	// Ensure nodes exist in DB with inputs
	if err := e.ensureNodes(ctx, execID, graph, exec.Inputs); err != nil {
		log.Printf("[Engine] Failed to ensure nodes: execution_id=%s, error=%v", execID, err)
		return err
	}

	// Reconcile DAG
	log.Printf("[Engine] Starting reconciliation: execution_id=%s", execID)
	err = e.reconcile(ctx, execID, graph)
	if err != nil {
		log.Printf("[Engine] Reconciliation failed: execution_id=%s, error=%v", execID, err)
	} else {
		log.Printf("[Engine] Reconciliation completed: execution_id=%s", execID)
	}
	return err
}

// ensureNodes ensures all DAG nodes exist in DB
func (e *Engine) ensureNodes(ctx context.Context, execID uuid.UUID, graph *dag.Graph, execInputs map[string]any) error {
	log.Printf("[Engine] Ensuring nodes exist: execution_id=%s, total_nodes=%d", execID, len(graph.Nodes))
	for id, node := range graph.Nodes {
		// Merge workflow execution inputs with node's 'with' clause
		// Execution inputs take precedence over node's 'with' clause (user-provided values override defaults)
		nodeInputs := make(map[string]any)

		// First, copy node's 'with' clause (defaults from workflow definition)
		if node.With != nil {
			for k, v := range node.With {
				nodeInputs[k] = v
			}
			log.Printf("[Engine] Node 'with' clause applied: execution_id=%s, node_id=%s, with_values=%v", execID, string(id), node.With)
		}

		// Then, override with execution inputs (user-provided values take precedence)
		for k, v := range execInputs {
			nodeInputs[k] = v
		}
		if len(execInputs) > 0 {
			log.Printf("[Engine] Execution inputs applied: execution_id=%s, node_id=%s, execution_inputs=%v", execID, string(id), execInputs)
		}

		log.Printf("[Engine] Final node inputs: execution_id=%s, node_id=%s, final_inputs=%v", execID, string(id), nodeInputs)

		n := &execution.ExecutionNode{
			ExecutionID:  execID,
			NodeID:       string(id),
			ExecutorType: node.Uses,
			Status:       execution.NodePending,
			MaxAttempts:  3, // default retries if needed
			Attempt:      0,
			Input:        nodeInputs,
		}
		if err := e.NodeStore.Upsert(ctx, n); err != nil {
			log.Printf("[Engine] Failed to upsert node: execution_id=%s, node_id=%s, error=%v", execID, string(id), err)
			return err
		}
		log.Printf("[Engine] Node ensured: execution_id=%s, node_id=%s, executor_type=%s, input_count=%d", execID, string(id), node.Uses, len(nodeInputs))
	}
	log.Printf("[Engine] All nodes ensured: execution_id=%s", execID)
	return nil
}

// reconcile loops through DAG and executes ready nodes
func (e *Engine) reconcile(ctx context.Context, execID uuid.UUID, graph *dag.Graph) error {
	iteration := 0
	for {
		iteration++
		log.Printf("[Engine] Reconciliation iteration %d: execution_id=%s", iteration, execID)

		// Load all nodes for execution
		nodes, err := e.NodeStore.ListByExecution(ctx, execID)
		if err != nil {
			log.Printf("[Engine] Failed to list nodes: execution_id=%s, error=%v", execID, err)
			return err
		}

		// Track progress
		progress := false
		pendingCount := 0
		runningCount := 0
		succeededCount := 0
		failedCount := 0
		retryingCount := 0

		for i := range nodes {
			node := &nodes[i]
			switch node.Status {
			case execution.NodePending:
				pendingCount++
			case execution.NodeRunning:
				runningCount++
			case execution.NodeSucceeded:
				succeededCount++
			case execution.NodeFailed:
				failedCount++
			case execution.NodeRetrying:
				retryingCount++
			}

			if node.Status != execution.NodePending && node.Status != execution.NodeRetrying {
				continue
			}

			dagNode := graph.Nodes[dag.NodeID(node.NodeID)]

			if !e.dependenciesDone(node, nodes, graph) {
				continue
			}

			if !shouldExecute(dagNode, nodes) {
				log.Printf("[Engine] Skipping node due to condition: execution_id=%s, node_id=%s",
					execID, node.NodeID)

				// Mark node as skipped
				_ = e.NodeStore.MarkSkipped(ctx, node.ExecutionID, node.NodeID, map[string]any{
					"skipped": true,
				})
				progress = true
				continue
			}

			progress = true
			log.Printf("[Engine] Executing node: execution_id=%s, node_id=%s", execID, node.NodeID)

			if err := e.executeNode(ctx, node, graph); err != nil {
				continue
			}
		}

		log.Printf("[Engine] Node status summary: execution_id=%s, pending=%d, running=%d, succeeded=%d, failed=%d, retrying=%d",
			execID, pendingCount, runningCount, succeededCount, failedCount, retryingCount)

		// Check if DAG is completed
		if e.isCompleted(nodes, execID) {
			log.Printf("[Engine] DAG completed: execution_id=%s", execID)
			return nil
		}

		if !progress {
			// Deadlock detected
			log.Printf("[Engine] Deadlock detected: execution_id=%s", execID)
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
		log.Printf("[Engine] Node not found in graph: node_id=%s", node.NodeID)
		return false // Node not found in graph
	}

	if len(graphNode.Depends) == 0 {
		return true // No dependencies
	}

	// Create a map of completed nodes for quick lookup
	completed := make(map[string]bool)
	for _, n := range allNodes {
		if n.Status == execution.NodeSucceeded {
			completed[n.NodeID] = true
		}
	}

	// Check if all dependencies are completed
	missingDeps := []string{}
	for _, dep := range graphNode.Depends {
		depID := string(dep)
		if !completed[depID] {
			missingDeps = append(missingDeps, depID)
		}
	}

	if len(missingDeps) > 0 {
		log.Printf("[Engine] Dependencies not ready: execution_id=%s, node_id=%s, missing_deps=%v",
			node.ExecutionID, node.NodeID, missingDeps)
		return false
	}

	log.Printf("[Engine] All dependencies satisfied: execution_id=%s, node_id=%s, dependency_count=%d",
		node.ExecutionID, node.NodeID, len(graphNode.Depends))
	return true
}

// resolveInputs resolves node inputs from previous node outputs and workflow inputs
func (e *Engine) resolveInputs(ctx context.Context, node *execution.ExecutionNode, graph *dag.Graph) (map[string]any, error) {
	dagNode := graph.Nodes[dag.NodeID(node.NodeID)]
	if dagNode == nil {
		return node.Input, nil
	}

	resolved := make(map[string]any)

	// Start with base inputs (from 'with' clause or execution inputs)
	for k, v := range node.Input {
		resolved[k] = v
	}

	// Resolve inputs from previous node outputs if 'inputs' field is defined
	if dagNode.Inputs != nil {
		// Load all nodes to get outputs from dependencies
		allNodes, err := e.NodeStore.ListByExecution(ctx, node.ExecutionID)
		if err != nil {
			return nil, err
		}

		// Create a map of node outputs
		nodeOutputs := make(map[string]map[string]any)
		for _, n := range allNodes {
			if n.Status == execution.NodeSucceeded && n.Output != nil {
				// Extract output from the stored output map
				// Output is stored as {"output": actualOutputMap}
				if wrappedOutput, ok := n.Output["output"]; ok {
					if outputMap, ok := wrappedOutput.(map[string]any); ok {
						nodeOutputs[n.NodeID] = outputMap
					}
				} else {
					// If not wrapped, use the output directly
					nodeOutputs[n.NodeID] = n.Output
				}
			}
		}

		// Resolve each input field
		for inputKey, inputValue := range dagNode.Inputs {
			if inputMap, ok := inputValue.(map[string]interface{}); ok {
				// Check if it's a from_node reference
				if fromNode, ok := inputMap["from_node"].(string); ok {
					if key, ok := inputMap["key"].(string); ok {
						// Get output from the referenced node
						if depOutputs, exists := nodeOutputs[fromNode]; exists {
							if value, found := depOutputs[key]; found {
								resolved[inputKey] = value
								log.Printf("[Engine] Resolved input %s from node %s.%s: %v",
									inputKey, fromNode, key, value)
							} else {
								return nil, fmt.Errorf("key '%s' not found in output of node '%s'", key, fromNode)
							}
						} else {
							return nil, fmt.Errorf("node '%s' output not found (node may not have completed)", fromNode)
						}
					}
				} else {
					// Not a from_node reference, use the value as-is
					resolved[inputKey] = inputValue
				}
			} else {
				// Simple value, use as-is
				resolved[inputKey] = inputValue
			}
		}
	}

	return resolved, nil
}

// executeNode executes a single node
func (e *Engine) executeNode(ctx context.Context, node *execution.ExecutionNode, graph *dag.Graph) error {
	node.Attempt++
	log.Printf("[Engine] Executing node: execution_id=%s, node_id=%s, attempt=%d/%d, executor_type=%s",
		node.ExecutionID, node.NodeID, node.Attempt, node.MaxAttempts, node.ExecutorType)

	// Resolve inputs from previous node outputs if needed
	resolvedInputs, err := e.resolveInputs(ctx, node, graph)
	if err != nil {
		log.Printf("[Engine] Failed to resolve inputs: execution_id=%s, node_id=%s, error=%v",
			node.ExecutionID, node.NodeID, err)
		return e.failOrRetry(ctx, node, err)
	}

	// Resolve module first before marking as running
	mod, err := e.ModuleReg.Resolve(ctx, "", node.ExecutorType)
	if err != nil {
		log.Printf("[Engine] Failed to resolve module: execution_id=%s, node_id=%s, executor_type=%s, error=%v",
			node.ExecutionID, node.NodeID, node.ExecutorType, err)
		return e.failOrRetry(ctx, node, err)
	}
	log.Printf("[Engine] Module resolved: execution_id=%s, node_id=%s, module_id=%s, runtime=%s",
		node.ExecutionID, node.NodeID, mod.ID, mod.Runtime)

	execImpl, err := executor.Get(mod.Runtime)
	if err != nil {
		log.Printf("[Engine] Failed to get executor: execution_id=%s, node_id=%s, runtime=%s, error=%v",
			node.ExecutionID, node.NodeID, mod.Runtime, err)
		return e.failOrRetry(ctx, node, err)
	}

	if err := e.NodeStore.MarkRunning(ctx, node.ExecutionID, node.NodeID); err != nil {
		log.Printf("[Engine] Failed to mark node as running: execution_id=%s, node_id=%s, error=%v",
			node.ExecutionID, node.NodeID, err)
		return err
	}

	log.Printf("[Engine] Executing node with executor: execution_id=%s, node_id=%s, runtime=%s, inputs=%v",
		node.ExecutionID, node.NodeID, mod.Runtime, resolvedInputs)
	out, err := execImpl.Execute(ctx, graph.Nodes[dag.NodeID(node.NodeID)], resolvedInputs)
	if err != nil {
		log.Printf("[Engine] Node execution failed: execution_id=%s, node_id=%s, error=%v",
			node.ExecutionID, node.NodeID, err)
		return e.failOrRetry(ctx, node, err)
	}

	log.Printf("[Engine] Node execution succeeded: execution_id=%s, node_id=%s", node.ExecutionID, node.NodeID)
	return e.NodeStore.MarkSucceeded(ctx, node.ExecutionID, node.NodeID, map[string]any{"output": out})
}

// failOrRetry handles retry logic or marks node failed
func (e *Engine) failOrRetry(ctx context.Context, node *execution.ExecutionNode, execErr error) error {
	if node.Attempt < node.MaxAttempts {
		log.Printf("[Engine] Retrying node: execution_id=%s, node_id=%s, attempt=%d/%d, error=%v",
			node.ExecutionID, node.NodeID, node.Attempt, node.MaxAttempts, execErr)
		// Use IncrementAttempt which properly sets status to RETRYING
		if err := e.NodeStore.IncrementAttempt(ctx, node.ExecutionID, node.NodeID); err != nil {
			log.Printf("[Engine] Failed to update node for retry: execution_id=%s, node_id=%s, error=%v",
				node.ExecutionID, node.NodeID, err)
			return err
		}
		// Update local node status for consistency
		node.Status = execution.NodeRetrying
		node.Attempt++
		return nil
	}

	log.Printf("[Engine] Node failed after max attempts: execution_id=%s, node_id=%s, attempts=%d, error=%v",
		node.ExecutionID, node.NodeID, node.Attempt, execErr)
	return e.NodeStore.MarkFailed(ctx, node.ExecutionID, node.NodeID, map[string]any{"message": execErr.Error()})
}

// isCompleted checks if all nodes have finished
func (e *Engine) isCompleted(nodes []execution.ExecutionNode, execID uuid.UUID) bool {
	allSucceeded := true
	hasRunning := false

	for _, node := range nodes {
		switch node.Status {
		case execution.NodeFailed:
			log.Printf("[Engine] Execution failed due to node failure: execution_id=%s, failed_node_id=%s", execID, node.NodeID)
			_ = e.ExecStore.MarkFailed(context.Background(), execID, map[string]any{"reason": "one or more nodes failed"})
			return true
		case execution.NodePending, execution.NodeRetrying:
			allSucceeded = false
		case execution.NodeSkipped:
			// Skipped nodes are considered as completed for execution completion check
			// They don't block execution completion
		case execution.NodeRunning:
			// Check if node has exceeded max attempts while still running (stuck state)
			if node.Attempt >= node.MaxAttempts {
				log.Printf("[Engine] Node stuck in running state after max attempts: execution_id=%s, node_id=%s, attempts=%d",
					execID, node.NodeID, node.Attempt)
				// Mark node as failed
				_ = e.NodeStore.MarkFailed(context.Background(), node.ExecutionID, node.NodeID,
					map[string]any{"reason": "node stuck in running state after max attempts"})
				_ = e.ExecStore.MarkFailed(context.Background(), execID, map[string]any{"reason": "one or more nodes failed"})
				return true
			}
			hasRunning = true
			allSucceeded = false
		}
	}

	// If we have running nodes but no progress, check for deadlock
	if hasRunning && allSucceeded {
		log.Printf("[Engine] Warning: nodes in running state but marked as succeeded: execution_id=%s", execID)
		allSucceeded = false
	}

	if allSucceeded {
		// Verify all nodes are either succeeded or skipped (both are considered completed)
		for _, node := range nodes {
			if node.Status != execution.NodeSucceeded && node.Status != execution.NodeSkipped {
				log.Printf("[Engine] Node not completed but execution marked complete: execution_id=%s, node_id=%s, status=%s",
					execID, node.NodeID, node.Status)
				return false
			}
		}

		log.Printf("[Engine] All nodes completed: execution_id=%s, total_nodes=%d", execID, len(nodes))
		// Collect outputs
		outputs := map[string]any{}
		for _, node := range nodes {
			outputs[node.NodeID] = node.Output
		}
		log.Printf("[Engine] Marking execution as completed: execution_id=%s, output_count=%d", execID, len(outputs))
		_ = e.ExecStore.MarkCompleted(context.Background(), execID, map[string]any{"outputs": outputs})
		return true
	}

	return false
}

func shouldExecute(
	node *dag.Node,
	allNodes []execution.ExecutionNode,
) bool {

	if node.When == nil {
		return true
	}

	for _, n := range allNodes {
		if n.NodeID == node.When.FromNode && n.Status == execution.NodeSucceeded {
			out := n.Output

			// unwrap {"output": {...}}
			if wrapped, ok := out["output"]; ok {
				if m, ok := wrapped.(map[string]any); ok {
					out = m
				}
			}

			val := out[node.When.Key]

			// Compare values, handling type conversions for string comparisons
			if node.When.Equals != nil {
				equals := node.When.Equals
				// Handle nil values
				if val == nil && equals == nil {
					log.Printf("[Engine] Condition check: node=%s, from_node=%s, key=%s, val=nil, equals=nil, result=true",
						node.ID, node.When.FromNode, node.When.Key)
					return true
				}
				if val == nil || equals == nil {
					log.Printf("[Engine] Condition check: node=%s, from_node=%s, key=%s, val=%v, equals=%v, result=false (nil mismatch)",
						node.ID, node.When.FromNode, node.When.Key, val, equals)
					return false
				}
				// Convert both to strings for comparison
				valStr := fmt.Sprintf("%v", val)
				equalsStr := fmt.Sprintf("%v", equals)
				result := valStr == equalsStr
				log.Printf("[Engine] Condition check: node=%s, from_node=%s, key=%s, val=%v (%s), equals=%v (%s), result=%v",
					node.ID, node.When.FromNode, node.When.Key, val, valStr, equals, equalsStr, result)
				return result
			}
			if node.When.NotEquals != nil {
				notEquals := node.When.NotEquals
				// Handle nil values
				if val == nil && notEquals == nil {
					log.Printf("[Engine] Condition check: node=%s, from_node=%s, key=%s, val=nil, not_equals=nil, result=false",
						node.ID, node.When.FromNode, node.When.Key)
					return false
				}
				if val == nil || notEquals == nil {
					log.Printf("[Engine] Condition check: node=%s, from_node=%s, key=%s, val=%v, not_equals=%v, result=true (nil mismatch)",
						node.ID, node.When.FromNode, node.When.Key, val, notEquals)
					return true
				}
				valStr := fmt.Sprintf("%v", val)
				notEqualsStr := fmt.Sprintf("%v", notEquals)
				result := valStr != notEqualsStr
				log.Printf("[Engine] Condition check: node=%s, from_node=%s, key=%s, val=%v (%s), not_equals=%v (%s), result=%v",
					node.ID, node.When.FromNode, node.When.Key, val, valStr, notEquals, notEqualsStr, result)
				return result
			}
		}
	}

	log.Printf("[Engine] Condition check failed: node=%s, from_node=%s not found or not succeeded", node.ID, node.When.FromNode)
	return false
}
