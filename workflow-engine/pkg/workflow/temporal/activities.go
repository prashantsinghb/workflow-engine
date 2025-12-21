package temporal

import (
	"context"
	"fmt"
	"log"

	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	wfregistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

// globals injected by worker
var (
	ModuleRegistry *registry.ModuleRegistry
	ExecutionStore execution.Store
	WorkflowStore  wfregistry.WorkflowStore
)

func SetModuleRegistry(m *registry.ModuleRegistry) {
	ModuleRegistry = m
}

func SetExecutionStore(s execution.Store) {
	ExecutionStore = s
}

func SetWorkflowStore(s wfregistry.WorkflowStore) {
	WorkflowStore = s
}

// --- helper to merge inputs for a node ---
func mergeNodeInputs(node *dag.Node, workflowInputs map[string]interface{}, stepOutputs map[string]map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// workflow-level inputs
	for k, v := range workflowInputs {
		result[k] = v
	}

	// outputs from dependent nodes
	for _, dep := range node.Depends {
		if depOut, ok := stepOutputs[string(dep)]; ok {
			for k, v := range depOut {
				result[k] = v
			}
		}
	}

	// node-level 'with' overrides
	for k, v := range node.With {
		result[k] = v
	}

	return result
}

// --- helper to extract workflow inputs safely from Temporal payloads ---
func extractWorkflowInputs(inputs map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	// Temporal SDK may pass payloads as "payloads" key
	if plRaw, ok := inputs["payloads"]; ok {
		if arr, ok := plRaw.([]interface{}); ok && len(arr) > 0 {
			// take last item, should be a map[string]interface{}
			if last, ok := arr[len(arr)-1].(map[string]interface{}); ok {
				for k, v := range last {
					result[k] = v
				}
			}
		}
	}

	// fallback: merge any top-level keys
	for k, v := range inputs {
		if k != "payloads" {
			result[k] = v
		}
	}

	return result
}

// --- NodeActivity executes a workflow DAG inside Temporal ---
func NodeActivity(
	ctx context.Context,
	projectID string,
	workflowID string,
	inputs map[string]interface{},
) (map[string]interface{}, error) {

	if WorkflowStore == nil {
		return nil, fmt.Errorf("workflow store not set")
	}
	if ModuleRegistry == nil {
		return nil, fmt.Errorf("module registry not set")
	}

	// extract workflow-level inputs properly
	wfInputs := extractWorkflowInputs(inputs)

	// Load workflow
	wf, err := WorkflowStore.Get(ctx, projectID, workflowID)
	if err != nil {
		return nil, err
	}

	graph := dag.Build(wf.Def)

	done := map[dag.NodeID]bool{}
	stepOutputs := map[string]map[string]interface{}{}

	for len(done) < len(graph.Nodes) {
		progress := false

		for id, node := range graph.Nodes {
			if done[id] {
				continue
			}
			if !isReady(node, done) {
				continue
			}

			// inject execution context
			actCtx := executor.WithProjectID(ctx, projectID)
			actCtx = executor.WithStepOutputs(actCtx, stepOutputs)

			mod, err := ModuleRegistry.GetModule(actCtx, projectID, node.Uses, "")
			if err != nil {
				return nil, err
			}

			execImpl, ok := executor.All()[mod.Runtime]
			if !ok {
				return nil, fmt.Errorf("executor not found: %s", mod.Runtime)
			}

			// Merge inputs for this node
			nodeInputs := mergeNodeInputs(node, wfInputs, stepOutputs)

			// log for debugging
			log.Printf("Executing node %s with inputs: %+v\n", id, nodeInputs)

			// Execute node
			out, err := execImpl.Execute(actCtx, node, nodeInputs)
			if err != nil {
				return nil, err
			}

			stepOutputs[string(id)] = out
			done[id] = true
			progress = true
		}

		if !progress {
			return nil, fmt.Errorf("deadlock detected in DAG")
		}
	}

	return stepOutputsToFlat(stepOutputs), nil
}

// --- helpers to mark execution status ---
func MarkExecutionSucceeded(
	ctx context.Context,
	executionID string,
	outputs map[string]interface{},
) error {
	if ExecutionStore == nil {
		return fmt.Errorf("execution store not set")
	}
	return ExecutionStore.MarkCompleted(ctx, executionID, outputs)
}

func MarkExecutionFailed(
	ctx context.Context,
	executionID string,
	errMsg string,
) error {
	if ExecutionStore == nil {
		return fmt.Errorf("execution store not set")
	}
	return ExecutionStore.MarkFailed(ctx, executionID, errMsg)
}

// --- helper to check if node dependencies are satisfied ---
func isReady(node *dag.Node, done map[dag.NodeID]bool) bool {
	for _, dep := range node.Depends {
		if !done[dep] {
			return false
		}
	}
	return true
}

// --- flatten step outputs ---
func stepOutputsToFlat(in map[string]map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
