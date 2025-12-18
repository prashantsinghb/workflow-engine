package workflow

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type ProjectStorage struct {
	mu         sync.RWMutex
	workflows  map[string]map[string]*WorkflowDefinition // project_id -> workflow_id -> workflow
	executions map[string]map[string]*Execution          // project_id -> execution_id -> execution
}

var store = &ProjectStorage{
	workflows:  make(map[string]map[string]*WorkflowDefinition),
	executions: make(map[string]map[string]*Execution),
}

// RegisterWorkflow stores a workflow under a project
func RegisterWorkflow(projectID string, def *WorkflowDefinition) (string, error) {
	if err := Validate(def); err != nil {
		return "", err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if store.workflows[projectID] == nil {
		store.workflows[projectID] = make(map[string]*WorkflowDefinition)
	}
	workflowID := uuid.NewString()
	store.workflows[projectID][workflowID] = def
	return workflowID, nil
}

// StartWorkflow creates an in-memory execution
func StartWorkflow(projectID, workflowID string, inputs map[string]string) (string, error) {
	store.mu.RLock()
	wf, ok := store.workflows[projectID][workflowID]
	store.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("workflow %s not found in project %s", workflowID, projectID)
	}

	exec := &Execution{
		ID:       uuid.NewString(),
		Workflow: wf,
		Inputs:   inputs,
		State:    RUNNING,
		Output:   make(map[string]string),
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if store.executions[projectID] == nil {
		store.executions[projectID] = make(map[string]*Execution)
	}
	store.executions[projectID][exec.ID] = exec

	// Dummy execution logic: mark as succeeded
	go func() {
		// Simulate work
		exec.State = SUCCEEDED
		exec.Output = map[string]string{"message": "Workflow executed successfully"}
	}()

	return exec.ID, nil
}

// GetExecution retrieves an execution
func GetExecution(projectID, executionID string) (*Execution, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if store.executions[projectID] == nil {
		return nil, errors.New("project not found")
	}
	exec, ok := store.executions[projectID][executionID]
	if !ok {
		return nil, errors.New("execution not found")
	}
	return exec, nil
}
