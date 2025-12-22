package execution

import (
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	ExecutionPending   ExecutionStatus = "PENDING"
	ExecutionRunning   ExecutionStatus = "RUNNING"
	ExecutionSucceeded ExecutionStatus = "SUCCEEDED"
	ExecutionFailed    ExecutionStatus = "FAILED"
	ExecutionCancelled ExecutionStatus = "CANCELLED"
	ExecutionPaused    ExecutionStatus = "PAUSED"
)

type NodeStatus string

const (
	NodePending   NodeStatus = "PENDING"
	NodeRunning   NodeStatus = "RUNNING"
	NodeSucceeded NodeStatus = "SUCCEEDED"
	NodeFailed    NodeStatus = "FAILED"
	NodeRetrying  NodeStatus = "RETRYING"
	NodeSkipped   NodeStatus = "SKIPPED"
)

type Execution struct {
	ID uuid.UUID

	ProjectID  string
	WorkflowID string
	Version    int

	ClientRequestID string
	TriggerType     string

	Status ExecutionStatus

	Inputs  map[string]any
	Outputs map[string]any
	Error   map[string]any

	StartedAt   *time.Time
	CompletedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExecutionNode struct {
	ID uuid.UUID

	ExecutionID uuid.UUID
	NodeID      string

	ExecutorType string
	Status       NodeStatus

	Attempt     int
	MaxAttempts int

	Input  map[string]any
	Output map[string]any
	Error  map[string]any

	StartedAt   *time.Time
	CompletedAt *time.Time
	DurationMs  *int64
}

type ExecutionEvent struct {
	ID uuid.UUID

	ExecutionID uuid.UUID
	NodeID      *string

	EventType string
	Message   string
	Payload   map[string]any

	CreatedAt time.Time
}

type ExecutionStats struct {
	TotalExecutions   int64
	RunningExecutions int64
	SuccessCount      int64
	FailedCount       int64
}
