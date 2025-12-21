package main

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/prashantsinghb/workflow-engine/pkg/config"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	wfregistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
)

const (
	TemporalAddr = "127.0.0.1:7233"
	TaskQueue    = "workflow-task-queue"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// ---- STORES ----
	execStore := execution.NewPostgresStore(db)
	workflowStore := wfregistry.NewPostgresWorkflowStore(db)

	// ---- REGISTRIES ----
	modulePg := registry.NewPostgresRegistry(db)
	moduleRegistry := registry.NewModuleRegistry(modulePg)

	// ---- TEMPORAL GLOBALS ----
	temporal.SetExecutionStore(execStore)
	temporal.SetWorkflowStore(workflowStore)
	temporal.SetModuleRegistry(moduleRegistry)

	// ---- EXECUTORS ----
	executor.Register("http", executor.NewHttpExecutor(moduleRegistry))
	executor.Register("noop", &executor.NoopExecutor{})

	seen := map[string]bool{}
	var mu sync.Mutex

	for {
		namespaces, _ := listNamespaces()
		for _, ns := range namespaces {
			if ns == "temporal-system" {
				continue
			}
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
	c, err := client.Dial(client.Options{HostPort: TemporalAddr})
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
	c, err := client.Dial(client.Options{
		HostPort:  TemporalAddr,
		Namespace: namespace,
	})
	if err != nil {
		return
	}
	defer c.Close()

	_ = temporal.StartWorker(c, TaskQueue)
}
