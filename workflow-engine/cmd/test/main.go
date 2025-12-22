package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	modreg "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	dag "github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	wfreg "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

var (
	stepRegistry   = wfreg.NewLocalStepRegistry()
	moduleRegistry *modreg.ModuleRegistry
)

func main() {
	ctx := context.Background()

	// Initialize DB
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/workflow?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	moduleRegistry = modreg.NewModuleRegistry(modreg.NewPostgresRegistry(db))

	// Initialize services and register steps
	if err := initServices(ctx); err != nil {
		log.Fatal(err)
	}

	// Build workflow DAG nodes
	nodes := []dag.Node{
		{ID: "1", Executor: "dnsworkflow.creatednsrecord", Uses: "service-a", With: map[string]interface{}{"domain": "example.com"}},
		{ID: "2", Executor: "iamworkflow.assignroles", Uses: "service-b", With: map[string]interface{}{"user_id": "u123", "roles": []string{"admin"}}},
	}

	// Execute workflow
	engine := NewEngine(stepRegistry, moduleRegistry)
	if err := engine.ExecuteWorkflow(ctx, nodes); err != nil {
		log.Fatal(err)
	}
}

// Engine executes a list of nodes using step and module registries
type Engine struct {
	stepRegistry   *wfreg.LocalStepRegistry
	moduleRegistry *modreg.ModuleRegistry
}

func NewEngine(stepRegistry *wfreg.LocalStepRegistry, moduleRegistry *modreg.ModuleRegistry) *Engine {
	return &Engine{stepRegistry, moduleRegistry}
}

func (e *Engine) ExecuteWorkflow(ctx context.Context, nodes []dag.Node) error {
	for _, node := range nodes {
		fmt.Println("Executing node:", node.Executor)

		// Local step execution
		step, err := e.stepRegistry.GetStep(node.Executor, "")
		if err == nil {
			out, err := step.Executor.Execute(ctx, nil, node.With)
			if err != nil {
				return err
			}
			fmt.Println("Output:", out)
			continue
		}

		// Remote step fallback
		mod, err := e.moduleRegistry.Resolve(ctx, "", node.Uses)
		if err != nil {
			return err
		}
		fmt.Println("Executing remote module:", mod.Name)
	}
	return nil
}
