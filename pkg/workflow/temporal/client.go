package temporal

import (
	"context"
	"log"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

var Client client.Client

func NewClient(addr, namespace string) {
	var err error
	Client, err = client.Dial(client.Options{
		HostPort:  addr,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("unable to create Temporal client: %v", err)
	}
}

func Close() {
	Client.Close()
}

func Describe(
	ctx context.Context,
	workflowID string,
	runID string,
) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return Client.DescribeWorkflowExecution(ctx, workflowID, runID)
}
