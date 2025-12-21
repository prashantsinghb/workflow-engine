package timeline

import (
	"time"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
)

type TimelineEventType string

const (
	TimelineExecutionStarted   TimelineEventType = "EXECUTION_STARTED"
	TimelineExecutionSucceeded TimelineEventType = "EXECUTION_SUCCEEDED"
	TimelineExecutionFailed    TimelineEventType = "EXECUTION_FAILED"

	TimelineNodeStarted   TimelineEventType = "NODE_STARTED"
	TimelineNodeSucceeded TimelineEventType = "NODE_SUCCEEDED"
	TimelineNodeFailed    TimelineEventType = "NODE_FAILED"
	TimelineNodeRetry     TimelineEventType = "NODE_RETRY"
)

type ExecutionTimelineEvent struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      TimelineEventType `json:"type"`

	NodeID   *string `json:"nodeId,omitempty"`
	Executor *string `json:"executor,omitempty"`

	Message    string         `json:"message,omitempty"`
	DurationMs *int64         `json:"durationMs,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type ExecutionTimeline struct {
	ExecutionID uuid.UUID `json:"executionId"`
	ProjectID   string    `json:"projectId"`
	WorkflowID  string    `json:"workflowId"`

	Status execution.ExecutionStatus `json:"status"`

	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	Events []ExecutionTimelineEvent `json:"timeline"`
}
