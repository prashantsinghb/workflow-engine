package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func StartWorker(c client.Client, taskQueue string) error {
	w := worker.New(c, taskQueue, worker.Options{})

	w.RegisterWorkflow(WorkflowExecution)
	w.RegisterActivity(NodeActivity)

	return w.Run(worker.InterruptCh())
}
