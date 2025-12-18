package workflow

import "errors"

var (
	ErrEmptyWorkflow     = errors.New("workflow must contain at least one node")
	ErrUnknownDependency = errors.New("unknown dependency")
	ErrCycleDetected     = errors.New("cycle detected in workflow")
	ErrWorkflowNotFound  = errors.New("workflow not found")
	ErrExecutionNotFound = errors.New("execution not found")
)
