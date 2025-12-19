package main

import (
	"log"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
	"go.temporal.io/sdk/worker"
)

func main() {
	temporal.NewClient("localhost:7233", "default")

	w := worker.New(
		temporal.Client,
		"workflow-task-queue",
		worker.Options{},
	)

	w.RegisterWorkflow(temporal.WorkflowExecution)
	w.RegisterActivity(temporal.NodeActivity)

	log.Println("Temporal worker started")
	log.Fatal(w.Run(worker.InterruptCh()))
}
