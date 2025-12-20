package execution

import "time"

type State string

const (
	StatePending   State = "PENDING"
	StateRunning   State = "RUNNING"
	StateSucceeded State = "SUCCEEDED"
	StateFailed    State = "FAILED"
	StateCancelled State = "CANCELLED"
)

type Execution struct {
	ID              string
	ProjectID       string
	WorkflowID      string
	ClientRequestID string

	TemporalWorkflowID string
	TemporalRunID      string

	State State
	Error string

	Inputs  map[string]interface{}
	Outputs map[string]interface{}

	StartedAt   *time.Time
	CompletedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
