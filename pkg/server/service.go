package server

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
)

type WorkflowServer struct {
	service.UnimplementedWorkflowServiceServer
	store *workflow.InMemoryStore
}

func NewWorkflowServer() *WorkflowServer {
	return &WorkflowServer{
		store: workflow.NewInMemoryStore(),
	}
}

func (s *WorkflowServer) ValidateWorkflow(ctx context.Context, req *service.ValidateWorkflowRequest) (*service.ValidateWorkflowResponse, error) {
	def, err := workflow.Parse([]byte(req.Workflow.Yaml))
	if err != nil {
		return &service.ValidateWorkflowResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	g := dag.Build(def)
	if err := dag.Validate(g); err != nil {
		return &service.ValidateWorkflowResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	return &service.ValidateWorkflowResponse{
		Valid: true,
	}, nil
}

func (s *WorkflowServer) RegisterWorkflow(ctx context.Context, req *service.RegisterWorkflowRequest) (*service.RegisterWorkflowResponse, error) {
	id, err := s.store.RegisterWorkflow(req.ProjectId, &workflow.WorkflowDefinition{
		Name:    req.Workflow.Name,
		Version: req.Workflow.Version,
		YAML:    req.Workflow.Yaml,
	})
	if err != nil {
		return nil, err
	}
	return &service.RegisterWorkflowResponse{WorkflowId: id}, nil
}

func (s *WorkflowServer) StartWorkflow(ctx context.Context, req *service.StartWorkflowRequest) (*service.StartWorkflowResponse, error) {
	id, err := s.store.StartWorkflow(req.ProjectId, req.WorkflowId, req.Inputs)
	if err != nil {
		return nil, err
	}
	return &service.StartWorkflowResponse{ExecutionId: id}, nil
}

func (s *WorkflowServer) GetExecution(ctx context.Context, req *service.GetExecutionRequest) (*service.GetExecutionResponse, error) {
	exec, err := s.store.GetExecution(req.ProjectId, req.ExecutionId)
	if err != nil {
		return nil, err
	}
	return &service.GetExecutionResponse{
		State:   service.ExecutionState(service.ExecutionState_value[string(exec.State)]),
		Outputs: exec.Output,
		Error:   exec.Error,
	}, nil
}
