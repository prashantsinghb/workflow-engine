package temporal

import (
	"context"
	"fmt"

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

			out, err := execImpl.Execute(actCtx, node, inputs)
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

func isReady(node *dag.Node, done map[dag.NodeID]bool) bool {
	for _, dep := range node.Depends {
		if !done[dep] {
			return false
		}
	}
	return true
}

func stepOutputsToFlat(in map[string]map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
