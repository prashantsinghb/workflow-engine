package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/api/enums/v1"
)

type ExecutionState string

const (
	StateRunning ExecutionState = "RUNNING"
	StateSuccess ExecutionState = "SUCCESS"
	StateFailed  ExecutionState = "FAILED"
)

type ExecutionInfo struct {
	ID        string
	State     ExecutionState
	StartTime int64
	EndTime   int64
	Error     string
}

func GetExecution(
	ctx context.Context,
	projectID string,
	executionID string,
) (*ExecutionInfo, error) {
	desc, err := Describe(ctx, executionID, "")
	if err != nil {
		return nil, fmt.Errorf("describe workflow: %w", err)
	}

	info := desc.WorkflowExecutionInfo
	state := mapTemporalState(info.Status)

	var errMsg string
	if info.Status == enums.WORKFLOW_EXECUTION_STATUS_FAILED ||
		info.Status == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT ||
		info.Status == enums.WORKFLOW_EXECUTION_STATUS_TERMINATED {
		wf := Client.GetWorkflow(ctx, executionID, "")
		var resultErr error
		err := wf.Get(ctx, &resultErr)
		if err != nil {
			errMsg = err.Error()
		}
	}

	return &ExecutionInfo{
		ID:        executionID,
		State:     state,
		StartTime: info.StartTime.AsTime().Unix(),
		EndTime:   info.CloseTime.AsTime().Unix(),
		Error:     errMsg,
	}, nil
}

func mapTemporalState(s enums.WorkflowExecutionStatus) ExecutionState {
	switch s {
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return StateRunning
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return StateSuccess
	default:
		return StateFailed
	}
}
