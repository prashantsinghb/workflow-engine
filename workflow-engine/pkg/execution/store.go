package execution

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	Executions() ExecutionStore
	Nodes() NodeStore
	Events() EventStore
}

type ExecutionStore interface {
	Create(ctx context.Context, exec *Execution) error

	Get(
		ctx context.Context,
		projectID string,
		executionID uuid.UUID,
	) (*Execution, error)

	GetByIdempotencyKey(
		ctx context.Context,
		projectID, workflowID, clientRequestID string,
	) (*Execution, error)

	MarkRunning(
		ctx context.Context,
		executionID uuid.UUID,
		runID string,
	) error

	MarkCompleted(
		ctx context.Context,
		executionID uuid.UUID,
		outputs map[string]any,
	) error

	MarkFailed(
		ctx context.Context,
		executionID uuid.UUID,
		err map[string]any,
	) error

	List(
		ctx context.Context,
		projectID, workflowID string,
	) ([]*Execution, error)

	ListRunning(ctx context.Context) ([]*Execution, error)

	GetStats(
		ctx context.Context,
		projectID string,
	) (*ExecutionStats, error)
}

type NodeStore interface {
	Upsert(ctx context.Context, node *ExecutionNode) error

	MarkRunning(
		ctx context.Context,
		executionID uuid.UUID,
		nodeID string,
	) error

	MarkSucceeded(
		ctx context.Context,
		executionID uuid.UUID,
		nodeID string,
		output map[string]any,
	) error

	MarkFailed(
		ctx context.Context,
		executionID uuid.UUID,
		nodeID string,
		err map[string]any,
	) error

	IncrementAttempt(
		ctx context.Context,
		executionID uuid.UUID,
		nodeID string,
	) error

	ListByExecution(
		ctx context.Context,
		executionID uuid.UUID,
	) ([]ExecutionNode, error)
}

type EventStore interface {
	Append(ctx context.Context, event *ExecutionEvent) error
	List(ctx context.Context, executionID uuid.UUID) ([]ExecutionEvent, error)
}
