package executor

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
)

type Executor interface {
	Execute(
		ctx context.Context,
		node *dag.Node,
		inputs map[string]interface{},
	) (map[string]interface{}, error)
}
