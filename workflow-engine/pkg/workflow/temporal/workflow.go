package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func WorkflowExecution(
	ctx workflow.Context,
	executionID string,
	projectID string,
	workflowID string,
	inputs map[string]interface{},
) error {

	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var outputs map[string]interface{}

	err := workflow.ExecuteActivity(
		ctx,
		NodeActivity,
		projectID,
		workflowID,
		inputs,
	).Get(ctx, &outputs)

	if err != nil {
		logger.Error("workflow failed", "error", err)

		_ = workflow.ExecuteActivity(
			ctx,
			MarkExecutionFailed,
			executionID,
			err.Error(),
		).Get(ctx, nil)

		return err
	}

	_ = workflow.ExecuteActivity(
		ctx,
		MarkExecutionSucceeded,
		executionID,
		outputs,
	).Get(ctx, nil)

	return nil
}
