package workflow

import (
	"errors"
)

// WorkflowDefinition holds raw YAML
type WorkflowDefinition struct {
	Name    string
	Version string
	YAML    string
}

// ExecutionState is used for in-memory execution
type ExecutionState string

const (
	PENDING   ExecutionState = "PENDING"
	RUNNING   ExecutionState = "RUNNING"
	SUCCEEDED ExecutionState = "SUCCEEDED"
	FAILED    ExecutionState = "FAILED"
)

// Execution represents a workflow run
type Execution struct {
	ID       string
	Workflow *WorkflowDefinition
	Inputs   map[string]string
	State    ExecutionState
	Output   map[string]string
	Error    string
}

// Simple DAG validation
func Validate(def *WorkflowDefinition) error {
	if def.Name == "" {
		return errors.New("workflow name required")
	}
	if def.YAML == "" {
		return errors.New("workflow YAML required")
	}
	// TODO: add actual DAG validation here
	return nil
}

// BuildGraph parses YAML (dummy for now)
func BuildGraph(def *WorkflowDefinition) (interface{}, error) {
	// TODO: parse YAML -> DAG structure
	return struct{}{}, nil
}
