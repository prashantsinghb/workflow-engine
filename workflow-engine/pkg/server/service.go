package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/structpb"

	service "github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	moduleregistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/parser"
	wfregistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/validation"
)

type WorkflowServer struct {
	service.UnimplementedWorkflowServiceServer

	execStore execution.Store
	wfStore   wfregistry.WorkflowStore
	modules   *moduleregistry.ModuleRegistry
	validator *validation.WorkflowValidator
}

func NewWorkflowService(
	execStore execution.Store,
	wfStore wfregistry.WorkflowStore,
	modules *moduleregistry.ModuleRegistry,
) *WorkflowServer {
	return &WorkflowServer{
		execStore: execStore,
		wfStore:   wfStore,
		modules:   modules,
		validator: validation.NewWorkflowValidator(),
	}
}

/* ---------------------- VALIDATE ---------------------- */

func (s *WorkflowServer) ValidateWorkflow(
	ctx context.Context,
	req *service.ValidateWorkflowRequest,
) (*service.ValidateWorkflowResponse, error) {

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

	if err := s.validator.Validate(ctx, &validation.Request{
		ProjectID:  req.ProjectId,
		Definition: def,
		Modules:    s.modules,
	}); err != nil {
		return &service.ValidateWorkflowResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	return &service.ValidateWorkflowResponse{Valid: true}, nil
}

/* ---------------------- REGISTER ---------------------- */

func (s *WorkflowServer) RegisterWorkflow(
	ctx context.Context,
	req *service.RegisterWorkflowRequest,
) (*service.RegisterWorkflowResponse, error) {

	if req.Workflow == nil {
		return nil, fmt.Errorf("workflow is required")
	}

	def, err := parser.ParseWorkflow([]byte(req.Workflow.Yaml))
	if err != nil {
		return nil, err
	}

	if err := s.validator.Validate(ctx, &validation.Request{
		ProjectID:  req.ProjectId,
		Definition: def,
		Modules:    s.modules,
	}); err != nil {
		return nil, err
	}

	wf := &wfregistry.Workflow{
		Name:      req.Workflow.Name,
		Version:   req.Workflow.Version,
		Yaml:      req.Workflow.Yaml,
		Def:       def,
		ProjectID: req.ProjectId,
	}

	id, err := s.wfStore.Register(ctx, req.ProjectId, wf)
	if err != nil {
		return nil, err
	}

	return &service.RegisterWorkflowResponse{
		WorkflowId: id,
	}, nil
}

/* ---------------------- LIST ---------------------- */

func (s *WorkflowServer) ListWorkflows(
	ctx context.Context,
	req *service.ListWorkflowsRequest,
) (*service.ListWorkflowsResponse, error) {

	workflows, err := s.wfStore.List(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	out := make([]*service.WorkflowInfo, 0, len(workflows))
	for _, wf := range workflows {
		out = append(out, &service.WorkflowInfo{
			Id:        wf.ID,
			Name:      wf.Name,
			Version:   wf.Version,
			ProjectId: wf.ProjectID,
		})
	}

	return &service.ListWorkflowsResponse{Workflows: out}, nil
}

/* ---------------------- GET ---------------------- */

func (s *WorkflowServer) GetWorkflow(
	ctx context.Context,
	req *service.GetWorkflowRequest,
) (*service.GetWorkflowResponse, error) {

	wf, err := s.wfStore.Get(ctx, req.ProjectId, req.WorkflowId)
	if err != nil {
		return nil, err
	}

	return &service.GetWorkflowResponse{
		Workflow: &service.WorkflowInfo{
			Id:        wf.ID,
			Name:      wf.Name,
			Version:   wf.Version,
			ProjectId: wf.ProjectID,
		},
		Yaml: wf.Yaml,
	}, nil
}

/* ---------------------- START ---------------------- */

func (s *WorkflowServer) StartWorkflow(
	ctx context.Context,
	req *service.StartWorkflowRequest,
) (*service.StartWorkflowResponse, error) {

	temporalWorkflowID := fmt.Sprintf(
		"%s:%s:%s",
		req.ProjectId,
		req.WorkflowId,
		req.ClientRequestId,
	)

	inputs := make(map[string]interface{}, len(req.Inputs))
	for k, v := range req.Inputs {
		if v != nil {
			inputs[k] = v.AsInterface()
		}
	}

	exec := &execution.Execution{
		ID:                 uuid.NewString(),
		ProjectID:          req.ProjectId,
		WorkflowID:         req.WorkflowId,
		ClientRequestID:    req.ClientRequestId,
		TemporalWorkflowID: temporalWorkflowID,
		State:              execution.StatePending,
		Inputs:             inputs,
	}

	err := s.execStore.CreateExecution(ctx, exec)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			existing, err := s.execStore.GetByIdempotencyKey(
				ctx,
				req.ProjectId,
				req.WorkflowId,
				req.ClientRequestId,
			)
			if err != nil {
				return nil, err
			}
			return &service.StartWorkflowResponse{
				ExecutionId: existing.ID,
				State:       string(existing.State),
			}, nil
		}
		return nil, err
	}

	tc, err := temporal.GetClientForProject(req.ProjectId)
	if err != nil {
		return nil, err
	}

	opts := client.StartWorkflowOptions{
		ID:        temporalWorkflowID,
		TaskQueue: "workflow-task-queue",
	}

	we, err := tc.Client.ExecuteWorkflow(
		ctx,
		opts,
		temporal.WorkflowExecution,
		exec.ID,
		req.ProjectId,
		req.WorkflowId,
		inputs,
	)
	if err != nil {
		return nil, err
	}

	if err := s.execStore.MarkRunning(ctx, exec.ID, we.GetRunID()); err != nil {
		return nil, err
	}

	return &service.StartWorkflowResponse{
		ExecutionId: exec.ID,
		State:       string(execution.StateRunning),
	}, nil
}

/* ---------------------- GET EXECUTION ---------------------- */

func (s *WorkflowServer) GetExecution(
	ctx context.Context,
	req *service.GetExecutionRequest,
) (*service.GetExecutionResponse, error) {

	exec, err := s.execStore.GetExecution(ctx, req.ProjectId, req.ExecutionId)
	if err != nil {
		return nil, err
	}

	var state service.ExecutionState
	switch exec.State {
	case execution.StatePending:
		state = service.ExecutionState_PENDING
	case execution.StateRunning:
		state = service.ExecutionState_RUNNING
	case execution.StateSucceeded:
		state = service.ExecutionState_SUCCESS
	case execution.StateFailed:
		state = service.ExecutionState_FAILED
	default:
		state = service.ExecutionState_EXECUTION_STATE_UNSPECIFIED
	}

	outputs := map[string]*structpb.Value{}
	for k, v := range exec.Outputs {
		val, _ := structpb.NewValue(v)
		outputs[k] = val
	}

	return &service.GetExecutionResponse{
		State:   state,
		Outputs: outputs,
		Error:   exec.Error,
	}, nil
}
