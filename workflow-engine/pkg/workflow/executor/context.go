package executor

import "context"

type ctxKey string

const projectIDKey ctxKey = "projectID"

type stepOutputsKeyType struct{}

var stepOutputsKey = stepOutputsKeyType{}

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}

func ProjectID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(projectIDKey).(string)
	return id, ok
}

// WithStepOutputs injects previous step outputs into context
func WithStepOutputs(
	ctx context.Context,
	outputs map[string]map[string]interface{},
) context.Context {
	return context.WithValue(ctx, stepOutputsKey, outputs)
}

// StepOutputs extracts step outputs from context
func StepOutputs(ctx context.Context) map[string]map[string]interface{} {
	if v := ctx.Value(stepOutputsKey); v != nil {
		if out, ok := v.(map[string]map[string]interface{}); ok {
			return out
		}
	}
	return nil
}
