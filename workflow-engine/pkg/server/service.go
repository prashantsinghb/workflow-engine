package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/structpb"

	service "github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	moduleregistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	wfexecution "github.com/prashantsinghb/workflow-engine/pkg/workflow/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/parser"
	wfregistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/validation"
)

type WorkflowServer struct {
	service.UnimplementedWorkflowServiceServer

	execStore execution.ExecutionStore
	wfStore   wfregistry.WorkflowStore
	modules   *moduleregistry.ModuleRegistry
	validator *validation.WorkflowValidator
	engine    *wfexecution.Engine
}

func NewWorkflowService(
	execStore execution.ExecutionStore,
	nodeStore execution.NodeStore,
	eventStore execution.EventStore,
	wfStore wfregistry.WorkflowStore,
	modules *moduleregistry.ModuleRegistry,
) *WorkflowServer {
	engine := &wfexecution.Engine{
		ExecStore:   execStore,
		NodeStore:   nodeStore,
		EventStore:  eventStore,
		WorkflowReg: wfStore,
		ModuleReg:   modules,
	}

	return &WorkflowServer{
		execStore: execStore,
		wfStore:   wfStore,
		modules:   modules,
		validator: validation.NewWorkflowValidator(),
		engine:    engine,
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
	inputs := make(map[string]interface{}, len(req.Inputs))
	for k, v := range req.Inputs {
		if v != nil {
			inputs[k] = v.AsInterface()
		}
	}

	exec := &execution.Execution{
		ID:              uuid.New(),
		ProjectID:       req.ProjectId,
		WorkflowID:      req.WorkflowId,
		ClientRequestID: req.ClientRequestId,
		Status:          execution.ExecutionPending,
		Inputs:          inputs,
	}

	if err := s.execStore.Create(ctx, exec); err != nil {
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
				ExecutionId: existing.ID.String(),
				State:       string(existing.Status),
			}, nil
		}

		return nil, err
	}

	go func() {
		bgCtx := context.Background()
		if err := s.engine.StartExecution(bgCtx, req.ProjectId, exec.ID); err != nil {
			_ = s.execStore.MarkFailed(bgCtx, exec.ID, map[string]any{
				"message": err.Error(),
			})
			return
		}
	}()

	return &service.StartWorkflowResponse{
		ExecutionId: exec.ID.String(),
		State:       string(execution.ExecutionPending),
	}, nil
}

/* ---------------------- GET EXECUTION ---------------------- */

func (s *WorkflowServer) GetExecution(
	ctx context.Context,
	req *service.GetExecutionRequest,
) (*service.GetExecutionResponse, error) {

	executionID, err := uuid.Parse(req.ExecutionId)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	exec, err := s.execStore.Get(ctx, req.ProjectId, executionID)
	if err != nil {
		return nil, err
	}

	var state service.ExecutionState
	switch exec.Status {
	case execution.ExecutionPending:
		state = service.ExecutionState_PENDING
	case execution.ExecutionRunning:
		state = service.ExecutionState_RUNNING
	case execution.ExecutionSucceeded:
		state = service.ExecutionState_SUCCESS
	case execution.ExecutionFailed:
		state = service.ExecutionState_FAILED
	default:
		state = service.ExecutionState_EXECUTION_STATE_UNSPECIFIED
	}

	inputs := map[string]*structpb.Value{}
	for k, v := range exec.Inputs {
		val, _ := structpb.NewValue(v)
		inputs[k] = val
	}

	outputs := map[string]*structpb.Value{}
	for k, v := range exec.Outputs {
		val, _ := structpb.NewValue(v)
		outputs[k] = val
	}

	var errorStr string
	if exec.Error != nil {
		// Try to extract message if present, otherwise marshal the whole map
		if msg, ok := exec.Error["message"].(string); ok {
			errorStr = msg
		} else {
			// Fallback to JSON marshaling
			errJSON, _ := json.Marshal(exec.Error)
			errorStr = string(errJSON)
		}
	}

	return &service.GetExecutionResponse{
		State:   state,
		Inputs:  inputs,
		Outputs: outputs,
		Error:   errorStr,
	}, nil
}

/* ---------------------- LIST EXECUTIONS ---------------------- */

func (s *WorkflowServer) ListExecutions(
	ctx context.Context,
	req *service.ListExecutionsRequest,
) (*service.ListExecutionsResponse, error) {

	execs, err := s.execStore.List(ctx, req.ProjectId, req.WorkflowId)
	if err != nil {
		return nil, err
	}

	res := &service.ListExecutionsResponse{
		Executions: make([]*service.ExecutionInfo, len(execs)),
	}

	// Fetch workflow names for all unique workflow IDs
	workflowNames := make(map[string]string)
	workflowIDs := make(map[string]bool)
	for _, e := range execs {
		if !workflowIDs[e.WorkflowID] {
			workflowIDs[e.WorkflowID] = true
		}
	}

	// Fetch workflow names
	for workflowID := range workflowIDs {
		wf, err := s.wfStore.Get(ctx, req.ProjectId, workflowID)
		if err == nil && wf != nil {
			workflowNames[workflowID] = wf.Name
		}
	}

	for i, e := range execs {
		workflowName := workflowNames[e.WorkflowID]
		if workflowName == "" {
			workflowName = e.WorkflowID // Fallback to ID if name not found
		}

		var errorStr string
		if e.Error != nil {
			// Try to extract message if present, otherwise marshal the whole map
			if msg, ok := e.Error["message"].(string); ok {
				errorStr = msg
			} else {
				// Fallback to JSON marshaling
				errJSON, _ := json.Marshal(e.Error)
				errorStr = string(errJSON)
			}
		}

		res.Executions[i] = &service.ExecutionInfo{
			Id:              e.ID.String(),
			WorkflowId:      e.WorkflowID,
			WorkflowName:    workflowName,
			ProjectId:       e.ProjectID,
			ClientRequestId: e.ClientRequestID,
			State:           string(e.Status),
			Error:           errorStr,
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
		TotalWorkflows:    totalWorkflows,
		TotalExecutions:   execStats.TotalExecutions,
		RunningExecutions: execStats.RunningExecutions,
		SuccessRate:       successRate,
	}, nil
}
