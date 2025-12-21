package executor

import "context"

type ctxKey string

const projectIDKey ctxKey = "projectID"

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}

func ProjectID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(projectIDKey).(string)
	return id, ok
}
