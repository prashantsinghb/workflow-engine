package temporal

import (
	"context"
	"fmt"

	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
)

var executionStoreForActivity execution.Store

func SetExecutionStoreForActivity(store execution.Store) {
	executionStoreForActivity = store
}

func NodeActivity(ctx context.Context, node *dag.Node, inputs map[string]interface{}) (map[string]interface{}, error) {
	execImpl, ok := executor.All()[node.Executor]
	if !ok {
		return nil, fmt.Errorf("executor not found: %s", node.Executor)
	}
	outputs, err := execImpl.Execute(ctx, node, inputs)
	if err != nil {
		return nil, err
	}
	if executionStoreForActivity != nil {
		_ = executionStoreForActivity.UpdateNodeOutputs(ctx, string(node.ID), outputs)
	}
	return outputs, nil
}
