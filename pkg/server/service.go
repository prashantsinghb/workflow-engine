package server

import (
	"context"
	"fmt"

	"github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/parser"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/structpb"
)

type WorkflowServer struct {
	service.UnimplementedWorkflowServiceServer
}

func (s *WorkflowServer) ValidateWorkflow(ctx context.Context, req *service.ValidateWorkflowRequest) (*service.ValidateWorkflowResponse, error) {
	if req.Workflow == nil {
		return &service.ValidateWorkflowResponse{
			Valid:  false,
			Errors: []string{"workflow is required"},
		}, nil
	}

	def, err := parser.ParseWorkflow([]byte(req.Workflow.Yaml))
	if err != nil {
		return &service.ValidateWorkflowResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	g := dag.Build(def)
	if err := dag.Validate(*g); err != nil {
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
	def, err := parser.ParseWorkflow([]byte(req.Workflow.Yaml))
	if err != nil {
		return nil, err
	}

	wf := &registry.Workflow{
		Name:    req.Workflow.Name,
		Version: req.Workflow.Version,
		Yaml:    req.Workflow.Yaml,
		Def:     def,
	}

	id, err := registry.Register(req.ProjectId, wf)
	if err != nil {
		return nil, err
	}

	return &service.RegisterWorkflowResponse{WorkflowId: id}, nil
}

func (s *WorkflowServer) StartWorkflow(ctx context.Context, req *service.StartWorkflowRequest) (*service.StartWorkflowResponse, error) {
	def, err := registry.Get(req.ProjectId, req.WorkflowId)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}
	g := dag.Build(def.Def)

	inputs := map[string]interface{}{}
	for k, v := range req.Inputs {
		inputs[k] = v.AsInterface()
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        req.ProjectId + "-" + req.WorkflowId,
		TaskQueue: "workflow-task-queue",
	}

	we, err := temporal.Client.ExecuteWorkflow(ctx, workflowOptions, temporal.WorkflowExecution, g, inputs)
	if err != nil {
		return nil, err
	}

	return &service.StartWorkflowResponse{
		ExecutionId: we.GetID(),
	}, nil
}

func (s *WorkflowServer) GetExecution(ctx context.Context, req *service.GetExecutionRequest) (*service.GetExecutionResponse, error) {
	info, err := temporal.GetExecution(
		ctx,
		req.ProjectId,
		req.ExecutionId,
	)
	if err != nil {
		return nil, err
	}

	return &service.GetExecutionResponse{
		State: service.ExecutionState(
			service.ExecutionState_value[string(info.State)],
		),
		Error: info.Error,
	}, nil
}

func structpbToInterface(v *structpb.Value) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	return v.AsInterface(), nil
}
