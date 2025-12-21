package registry

import (
	"context"
)

type WorkflowStore interface {
	Register(ctx context.Context, projectID string, wf *Workflow) (string, error)
	Get(ctx context.Context, projectID string, workflowID string) (*Workflow, error)
	List(ctx context.Context, projectID string) ([]*Workflow, error)
}
