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

	// Start workflow asynchronously to avoid blocking HTTP request
	// Use a goroutine to start the workflow in the background
	go func() {
		bgCtx := context.Background()
		
		tc, err := temporal.GetClientForProject(req.ProjectId)
		if err != nil {
			// Mark execution as failed if we can't connect to Temporal
			_ = s.execStore.MarkFailed(bgCtx, exec.ID, fmt.Sprintf("Failed to connect to Temporal: %v", err))
			return
		}

		opts := client.StartWorkflowOptions{
			ID:        temporalWorkflowID,
			TaskQueue: "workflow-task-queue",
		}

		we, err := tc.Client.ExecuteWorkflow(
			bgCtx,
			opts,
			temporal.WorkflowExecution,
			exec.ID,
			req.ProjectId,
			req.WorkflowId,
			inputs,
		)
		if err != nil {
			// Mark execution as failed if workflow start fails
			_ = s.execStore.MarkFailed(bgCtx, exec.ID, fmt.Sprintf("Failed to start workflow: %v", err))
			return
		}

		// Mark execution as running once workflow is started
		_ = s.execStore.MarkRunning(bgCtx, exec.ID, we.GetRunID())
	}()

	// Return immediately with PENDING state
	// The workflow will be started asynchronously and state will update to RUNNING
	return &service.StartWorkflowResponse{
		ExecutionId: exec.ID,
		State:       string(execution.StatePending),
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

/* ---------------------- LIST EXECUTIONS ---------------------- */

func (s *WorkflowServer) ListExecutions(
	ctx context.Context,
	req *service.ListExecutionsRequest,
) (*service.ListExecutionsResponse, error) {

	execs, err := s.execStore.ListExecutions(ctx, req.ProjectId, req.WorkflowId)
	if err != nil {
		return nil, err
	}

	res := &service.ListExecutionsResponse{
		Executions: make([]*service.ExecutionInfo, len(execs)),
	}

	for i, e := range execs {
		res.Executions[i] = &service.ExecutionInfo{
			Id:              e.ID,
			WorkflowId:      e.WorkflowID,
			ProjectId:       e.ProjectID,
			ClientRequestId: e.ClientRequestID,
			State:           string(e.State),
			Error:           e.Error,
		}
	}

	return res, nil
}

/* ---------------------- DASHBOARD STATS ---------------------- */

func (s *WorkflowServer) GetDashboardStats(
	ctx context.Context,
	req *service.GetDashboardStatsRequest,
) (*service.GetDashboardStatsResponse, error) {

	// Get workflow count
	totalWorkflows, err := s.wfStore.Count(ctx, req.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to count workflows: %w", err)
	}

	// Get execution stats
	execStats, err := s.execStore.GetStats(ctx, req.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution stats: %w", err)
	}

	// Calculate success rate (only for completed executions)
	var successRate float64
	completedCount := execStats.SuccessCount + execStats.FailedCount
	if completedCount > 0 {
		successRate = (float64(execStats.SuccessCount) / float64(completedCount)) * 100.0
	}

	return &service.GetDashboardStatsResponse{
		TotalWorkflows:   totalWorkflows,
		TotalExecutions:  execStats.TotalExecutions,
		RunningExecutions: execStats.RunningExecutions,
		SuccessRate:      successRate,
	}, nil
}
