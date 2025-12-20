package registry

import (
	"sync"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
)

type Workflow struct {
	ID      string
	Name    string
	Version string
	Yaml    string
	Def     *api.Definition
}

var (
	mu        sync.Mutex
	workflows = make(map[string]map[string]*Workflow)
)

func Register(projectID string, workflow *Workflow) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if workflows[projectID] == nil {
		workflows[projectID] = make(map[string]*Workflow)
	}

	workflow.ID = uuid.NewString()
	workflows[projectID][workflow.ID] = workflow
	return workflow.ID, nil
}

func Get(projectID, workflowID string) (*Workflow, error) {
	mu.Lock()
	defer mu.Unlock()

	if proj, ok := workflows[projectID]; ok {
		if wf, ok := proj[workflowID]; ok {
			return wf, nil
		}
	}
	return nil, workflow.ErrWorkflowNotFound
}
