package temporal

import (
	"context"
	"fmt"

	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
)

// moduleRegistry must be set before starting worker
var ModuleRegistry *registry.ModuleRegistry
var ExecutionStore execution.Store

func SetModuleRegistry(mr *registry.ModuleRegistry) {
	ModuleRegistry = mr
}

func SetExecutionStore(store execution.Store) {
	ExecutionStore = store
}

// NodeActivity is called by Temporal for each DAG node
func NodeActivity(ctx context.Context, projectID string, node *dag.Node, inputs map[string]interface{}) (map[string]interface{}, error) {
	if ModuleRegistry == nil {
		return nil, fmt.Errorf("ModuleRegistry is not set")
	}

	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}

	// Set projectID in context for executors
	ctx = executor.WithProjectID(ctx, projectID)

	// 1️⃣ Resolve module
	mod, err := ModuleRegistry.GetModule(ctx, projectID, node.Uses, "")
	if err != nil {
		return nil, fmt.Errorf("module %s not found: %w", node.Uses, err)
	}

	// 2️⃣ Resolve executor from module runtime
	execImpl, ok := executor.All()[mod.Runtime]
	if !ok {
		return nil, fmt.Errorf("executor not found: %s", mod.Runtime)
	}

	// 3️⃣ Execute
	outputs, err := execImpl.Execute(ctx, node, inputs)
	if err != nil {
		return nil, err
	}

	// 4️⃣ Persist outputs if execution store is set
	if ExecutionStore != nil {
		_ = ExecutionStore.UpdateNodeOutputs(ctx, string(node.ID), outputs)
	}

	return outputs, nil
}
