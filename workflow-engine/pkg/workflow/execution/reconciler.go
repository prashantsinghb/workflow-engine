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

	// TODO: Reconcile running executions with Temporal if needed
	// This requires Temporal workflow IDs to be stored, which are currently removed
	for _, e := range execs {
		// Skip reconciliation for now since TemporalWorkflowID is removed
		_ = e
	}
}
