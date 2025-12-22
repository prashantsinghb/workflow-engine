package execution

import (
	"context"
	"log"
)

func Reconcile(ctx context.Context, executionContext *Context) {
	execs, err := executionContext.Store.Executions().ListRunning(ctx)
	if err != nil {
		log.Println("Reconcile: failed to list running executions:", err)
		return
	}

	for _, e := range execs {
		we := executionContext.Temporal.GetWorkflow(ctx, e.TemporalWorkflowID, "")
		var status string
		err := we.Get(ctx, &status)
		if err != nil {
			executionContext.Store.Executions().MarkFailed(ctx, e.ID, map[string]any{"message": err.Error()})
		} else if status == "COMPLETED" {
			executionContext.Store.Executions().MarkCompleted(ctx, e.ID, nil)
		}
	}
}
