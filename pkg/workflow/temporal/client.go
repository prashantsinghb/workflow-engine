package temporal

import (
	"context"
	"sync"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	mu      sync.Mutex
	clients = map[string]*Client{} // projectID → Client
)

type Client struct {
	Client client.Client
}

func GetClientForProject(projectID string) (*Client, error) {
	mu.Lock()
	defer mu.Unlock()

	if c, ok := clients[projectID]; ok {
		return c, nil
	}

	// Connect to Temporal service
	cli, err := client.Dial(client.Options{
		Namespace: projectID,
	})
	if err != nil {
		return nil, err
	}

	c := &Client{Client: cli}

	// Ensure namespace exists
	if err := createNamespaceIfNotExists(c, projectID); err != nil {
		cli.Close()
		return nil, err
	}

	clients[projectID] = c
	return c, nil
}

func (c *Client) Close() {
	c.Client.Close()
}

func Describe(
	ctx context.Context,
	c *Client,
	workflowID string,
	runID string,
) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Client.DescribeWorkflowExecution(ctx, workflowID, runID)
}

func createNamespaceIfNotExists(c *Client, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	retention := durationpb.New(24 * time.Hour)

	_, err := c.Client.WorkflowService().RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespace,
		WorkflowExecutionRetentionPeriod: retention,
	})
	if err != nil {
		// Ignore "already exists" error
		if !isNamespaceAlreadyExistsError(err) {
			return err
		}
	}
	return nil
}

// isNamespaceAlreadyExistsError checks gRPC error type for existing namespace
func isNamespaceAlreadyExistsError(err error) bool {
	return err != nil && err.Error() == "Namespace already exists."
}
