package main

import (
	"context"
	"log"
	"sync"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
)

const (
	TemporalAddr = "localhost:7233"
	TaskQueue    = "workflow-task-queue"
)

func main() {
	log.Println("Starting Temporal worker (dynamic namespaces)")

	// Track which namespaces already have workers
	seen := map[string]bool{}
	var mu sync.Mutex

	for {
		namespaces, err := listNamespaces()
		if err != nil {
			log.Printf("failed to list namespaces: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, ns := range namespaces {
			mu.Lock()
			if seen[ns] {
				mu.Unlock()
				continue
			}
			seen[ns] = true
			mu.Unlock()

			go startWorker(ns)
		}

		time.Sleep(30 * time.Second)
	}
}

func listNamespaces() ([]string, error) {
	c, err := client.Dial(client.Options{
		HostPort: TemporalAddr,
	})
	if err != nil {
		return nil, err
	}
	defer c.Close()

	resp, err := c.WorkflowService().ListNamespaces(context.Background(), &workflowservice.ListNamespacesRequest{})
	if err != nil {
		return nil, err
	}

	var out []string
	for _, ns := range resp.Namespaces {
		out = append(out, ns.NamespaceInfo.Name)
	}
	return out, nil
}

func startWorker(namespace string) {
	log.Printf("Starting worker for namespace: %s", namespace)

	c, err := client.Dial(client.Options{
		HostPort:  TemporalAddr,
		Namespace: namespace,
	})
	if err != nil {
		log.Printf("client error for %s: %v", namespace, err)
		return
	}

	// Initialize module registry (use DB if persistence is required)
	moduleRegistry := registry.NewModuleRegistry(nil)
	temporal.SetModuleRegistry(moduleRegistry)

	// Register global executors
	executor.Register("http", executor.NewHttpExecutor(moduleRegistry))
	executor.Register("noop", &executor.NoopExecutor{})
	// TODO: container executor
	// executor.Register("container", executor.NewContainerExecutor(moduleRegistry))

	w := worker.New(c, TaskQueue, worker.Options{})

	// Register workflow & activities
	w.RegisterWorkflow(temporal.WorkflowExecution)
	w.RegisterActivity(temporal.NodeActivity)

	log.Printf("Worker running for namespace: %s", namespace)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Printf("worker failed for %s: %v", namespace, err)
	}
}
