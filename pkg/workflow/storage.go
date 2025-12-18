package workflow

import (
	"sync"

	"github.com/google/uuid"
)

type InMemoryStore struct {
	mu         sync.RWMutex
	workflows  map[string]map[string]*WorkflowDefinition // project_id -> workflow_id -> workflow
	executions map[string]map[string]*Execution          // project_id -> execution_id -> execution
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		workflows:  make(map[string]map[string]*WorkflowDefinition),
		executions: make(map[string]map[string]*Execution),
	}
}

var store = &InMemoryStore{
	workflows:  make(map[string]map[string]*WorkflowDefinition),
	executions: make(map[string]map[string]*Execution),
}

func (s *InMemoryStore) RegisterWorkflow(
	projectID string,
	def *WorkflowDefinition,
) (string, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.workflows[projectID] == nil {
		s.workflows[projectID] = make(map[string]*WorkflowDefinition)
	}

	id := uuid.NewString()
	s.workflows[projectID][id] = def

	return id, nil
}

func (s *InMemoryStore) StartWorkflow(
	projectID, workflowID string,
	inputs map[string]string,
) (string, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	workflow, ok := s.workflows[projectID][workflowID]
	if !ok {
		return "", ErrWorkflowNotFound
	}

	execID := uuid.NewString()
	exec := &Execution{
		ID:       execID,
		Workflow: workflow,
		State:    RUNNING,
		Inputs:   inputs,
		Output:   map[string]string{},
	}

	if s.executions[projectID] == nil {
		s.executions[projectID] = make(map[string]*Execution)
	}
	s.executions[projectID][execID] = exec

	// For now: immediately succeed
	exec.State = SUCCEEDED

	return execID, nil
}

func (s *InMemoryStore) GetExecution(
	projectID, executionID string,
) (*Execution, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	exec, ok := s.executions[projectID][executionID]
	if !ok {
		return nil, ErrExecutionNotFound
	}
	return exec, nil
}
