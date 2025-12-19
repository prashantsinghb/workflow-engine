package temporal

import (
	"context"
	"fmt"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
)

func NodeActivity(ctx context.Context, node *dag.Node, inputs map[string]interface{}) (map[string]interface{}, error) {
	execImpl, ok := executor.All()[node.Executor]
	if !ok {
		return nil, fmt.Errorf("executor not found: %s", node.Executor)
	}
	return execImpl.Execute(ctx, node, inputs)
}
