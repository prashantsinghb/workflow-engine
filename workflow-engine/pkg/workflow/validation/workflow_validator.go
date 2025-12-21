package validation

import (
	"context"
	"fmt"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
)

type WorkflowValidator struct{}

func NewWorkflowValidator() *WorkflowValidator {
	return &WorkflowValidator{}
}

func (v *WorkflowValidator) Validate(
	ctx context.Context,
	req *Request,
) error {

	if req.Definition == nil {
		return fmt.Errorf("workflow definition is nil")
	}

	if len(req.Definition.Nodes) == 0 {
		return fmt.Errorf("workflow must contain at least one node")
	}

	if err := v.validateDAG(req.Definition); err != nil {
		return err
	}

	if err := v.validateModules(ctx, req); err != nil {
		return err
	}

	// if err := v.validateExecutors(req.Definition); err != nil {
	// 	return err
	// }

	return nil
}

func (v *WorkflowValidator) validateDAG(def *api.Definition) error {
	g := dag.Build(def)
	if err := dag.Validate(*g); err != nil {
		return fmt.Errorf("dag validation failed: %w", err)
	}
	return nil
}

func (v *WorkflowValidator) validateModules(
	ctx context.Context,
	req *Request,
) error {

	for _, node := range req.Definition.Nodes {
		_, err := req.Modules.Resolve(
			ctx,
			req.ProjectID,
			node.Uses,
		)
		if err != nil {
			return fmt.Errorf(
				"node references unknown module: %s",
				node.Uses,
			)
		}
	}
	return nil
}

// func (v *WorkflowValidator) validateExecutors(def *api.Definition) error {
// 	for _, node := range def.Nodes {
// 		if !executor.Exists(node.Type) {
// 			return fmt.Errorf("unknown executor: %s", node.Type)
// 		}
// 	}
// 	return nil
// }
