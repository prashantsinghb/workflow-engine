package server

import (
	"context"
	"fmt"

	"github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/parser"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
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
	inputs := make(map[string]interface{}, len(req.Inputs))
	for k, v := range req.Inputs {
		val, err := structpbToInterface(v)
		if err != nil {
			return nil, fmt.Errorf("invalid input %s: %w", k, err)
		}
		inputs[k] = val
	}

	execID, err := execution.Start(req.ProjectId, req.WorkflowId, inputs)
	if err != nil {
		return nil, err
	}
	wf, _ := registry.Get(req.ProjectId, req.WorkflowId)
	for name, node := range wf.Def.Nodes {
		output, err := executor.RunNode(node, inputs)
		if err != nil {
			return &service.StartWorkflowResponse{ExecutionId: execID}, fmt.Errorf("node %s failed: %w", name, err)
		}
		_ = output
	}

	return &service.StartWorkflowResponse{ExecutionId: execID}, nil
}

func (s *WorkflowServer) GetExecution(ctx context.Context, req *service.GetExecutionRequest) (*service.GetExecutionResponse, error) {
	exec, err := execution.GetExecution(req.ProjectId, req.ExecutionId)
	if err != nil {
		return nil, err
	}
	state := service.ExecutionState(service.ExecutionState_value[string(exec.State)])
	outputs := make(map[string]*structpb.Value, len(exec.Output))
	for k, v := range exec.Output {
		val, err := structpb.NewValue(v)
		if err != nil {
			return nil, err
		}
		outputs[k] = val
	}
	return &service.GetExecutionResponse{
		State:   state,
		Outputs: outputs,
		Error:   exec.Error,
	}, nil
}

func structpbToInterface(v *structpb.Value) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	return v.AsInterface(), nil
}
