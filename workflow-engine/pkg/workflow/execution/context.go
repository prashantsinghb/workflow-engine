package execution

import (
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	wfRegistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
	"go.temporal.io/sdk/client"
)

type Context struct {
	// Persistence
	Store execution.Store

	// Step resolution
	Modules registry.ModuleRegistry

	Workflow wfRegistry.WorkflowStore

	Temporal client.Client
}
