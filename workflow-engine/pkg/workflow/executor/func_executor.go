package executor

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
)

// FuncExecutor wraps a simple Go function into an Executor
type FuncExecutor struct {
	fn func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

func NewFuncExecutor(fn func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)) *FuncExecutor {
	return &FuncExecutor{fn: fn}
}

// Execute implements Executor interface
func (f *FuncExecutor) Execute(ctx context.Context, node *dag.Node, inputs map[string]interface{}) (map[string]interface{}, error) {
	return f.fn(ctx, inputs)
}
