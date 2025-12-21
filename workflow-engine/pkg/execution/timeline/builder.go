package timeline

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
)

type TimelineBuilder struct {
	executions execution.ExecutionStore
	nodes      execution.NodeStore
	events     execution.EventStore
}

func NewTimelineBuilder(store execution.Store) *TimelineBuilder {
	return &TimelineBuilder{
		executions: store.Executions(),
		nodes:      store.Nodes(),
		events:     store.Events(),
	}
}

func (b *TimelineBuilder) Build(
	ctx context.Context,
	projectID string,
	executionID uuid.UUID,
) (*ExecutionTimeline, error) {

	exec, err := b.executions.Get(ctx, projectID, executionID)
	if err != nil {
		return nil, err
	}

	nodes, err := b.nodes.ListByExecution(ctx, executionID)
	if err != nil {
		return nil, err
	}

	events, err := b.events.List(ctx, executionID)
	if err != nil {
		return nil, err
	}

	var timeline []ExecutionTimelineEvent

	// 1. Execution started
	if exec.StartedAt != nil {
		timeline = append(timeline, ExecutionTimelineEvent{
			Timestamp: *exec.StartedAt,
			Type:      TimelineExecutionStarted,
			Message:   "Execution started",
		})
	}

	// 2. Node events
	for _, n := range nodes {
		if n.StartedAt != nil {
			timeline = append(timeline, ExecutionTimelineEvent{
				Timestamp: *n.StartedAt,
				Type:      TimelineNodeStarted,
				NodeID:    &n.NodeID,
				Executor:  &n.ExecutorType,
			})
		}

		if n.CompletedAt != nil {
			eventType := TimelineNodeSucceeded
			if n.Status == execution.NodeFailed {
				eventType = TimelineNodeFailed
			}

			timeline = append(timeline, ExecutionTimelineEvent{
				Timestamp:  *n.CompletedAt,
				Type:       eventType,
				NodeID:     &n.NodeID,
				DurationMs: n.DurationMs,
				Payload:    n.Error,
			})
		}
	}

	// 3. Explicit events (retries, pauses, etc.)
	for _, e := range events {
		timeline = append(timeline, ExecutionTimelineEvent{
			Timestamp: e.CreatedAt,
			Type:      TimelineEventType(e.EventType),
			NodeID:    e.NodeID,
			Message:   e.Message,
			Payload:   e.Payload,
		})
	}

	// 4. Execution end
	if exec.CompletedAt != nil {
		endType := TimelineExecutionSucceeded
		if exec.Status == execution.ExecutionFailed {
			endType = TimelineExecutionFailed
		}

		timeline = append(timeline, ExecutionTimelineEvent{
			Timestamp: *exec.CompletedAt,
			Type:      endType,
		})
	}

	// 5. Sort chronologically
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp.Before(timeline[j].Timestamp)
	})

	return &ExecutionTimeline{
		ExecutionID: exec.ID,
		ProjectID:   exec.ProjectID,
		WorkflowID:  exec.WorkflowID,
		Status:      exec.Status,
		StartedAt:   exec.StartedAt,
		CompletedAt: exec.CompletedAt,
		Events:      timeline,
	}, nil
}
