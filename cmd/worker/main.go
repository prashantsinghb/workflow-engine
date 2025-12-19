package main

import (
	"context"
	"log"
	"sync"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
)

const (
	TemporalAddr = "localhost:7233"
	TaskQueue    = "workflow-task-queue"
)

func main() {
	log.Println("Starting Temporal worker (dynamic namespaces)")

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

	resp, err := c.WorkflowService().ListNamespaces(
		context.Background(),
		&workflowservice.ListNamespacesRequest{},
	)
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

	if err := temporal.StartWorker(c, TaskQueue); err != nil {
		log.Printf("worker failed for %s: %v", namespace, err)
	}
}
