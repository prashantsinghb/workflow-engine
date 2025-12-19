package execution

import (
	"fmt"
	"sync"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

type ExecutionState string

const (
	StateRunning ExecutionState = "RUNNING"
	StateSuccess ExecutionState = "SUCCESS"
	StateFailed  ExecutionState = "FAILED"
)

type Execution struct {
	ID       string
	Workflow *registry.Workflow
	State    ExecutionState
	Output   map[string]interface{}
	Error    string
}

var (
	mu         sync.Mutex
	executions = make(map[string]map[string]*Execution)
	counter    int
)

func Start(projectID, workflowID string, inputs map[string]interface{}) (string, error) {
	wf, err := registry.Get(projectID, workflowID)
	if err != nil {
		return "", err
	}

	mu.Lock()
	defer mu.Unlock()
	counter++
	execID := workflowID + "-exec-" + fmt.Sprint(counter)

	exec := &Execution{
		ID:       execID,
		Workflow: wf,
		State:    StateRunning,
		Output:   make(map[string]interface{}),
	}
	if executions[projectID] == nil {
		executions[projectID] = make(map[string]*Execution)
	}
	executions[projectID][execID] = exec
	return execID, nil
}

func GetExecution(projectID, execID string) (*Execution, error) {
	mu.Lock()
	defer mu.Unlock()
	if proj, ok := executions[projectID]; ok {
		if exec, ok := proj[execID]; ok {
			return exec, nil
		}
	}
	return nil, workflow.ErrExecutionNotFound
}
