package execution

import "context"

type Store interface {
	CreateExecution(ctx context.Context, exec *Execution) error
	GetExecution(ctx context.Context, projectID, executionID string) (*Execution, error)
	GetByIdempotencyKey(ctx context.Context, projectID string, workflowID string, clientRequestID string) (*Execution, error)
	MarkRunning(ctx context.Context, executionID, runID string) error
	ListRunningExecutions(ctx context.Context) ([]*Execution, error)
	MarkCompleted(ctx context.Context, executionID string, outputs map[string]interface{}) error
	MarkFailed(ctx context.Context, executionID string, errMsg string) error
	UpdateNodeOutputs(ctx context.Context, nodeID string, outputs map[string]interface{}) error
}
