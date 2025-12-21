package execution

import (
	"context"
	"log"

	"github.com/prashantsinghb/workflow-engine/pkg/execution"

	"go.temporal.io/sdk/client"
)

func Reconcile(ctx context.Context, store execution.Store, temporal client.Client) {
	execs, err := store.ListRunningExecutions(ctx)
	if err != nil {
		log.Println("Reconcile: failed to list running executions:", err)
		return
	}

	for _, e := range execs {
		we := temporal.GetWorkflow(ctx, e.TemporalWorkflowID, "")
		var status string
		err := we.Get(ctx, &status)
		if err != nil {
			store.MarkFailed(ctx, e.ID, err.Error())
		} else if status == "COMPLETED" {
			store.MarkCompleted(ctx, e.ID, nil)
		}
	}
}
