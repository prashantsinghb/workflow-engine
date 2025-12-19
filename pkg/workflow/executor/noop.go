package executor

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
)

type NoopExecutor struct{}

func (n *NoopExecutor) Execute(
	ctx context.Context,
	node *dag.Node,
	inputs map[string]interface{},
) (map[string]interface{}, error) {
	return map[string]interface{}{
		"node_id": string(node.ID),
	}, nil
}
