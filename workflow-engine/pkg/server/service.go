package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/parser"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/temporal"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/structpb"
)

type WorkflowServer struct {
	service.UnimplementedWorkflowServiceServer
	execStore execution.Store
}

func NewWorkflowService(execStore execution.Store) *WorkflowServer {
	return &WorkflowServer{
		execStore: execStore,
	}
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

func (s *WorkflowServer) GetExecution(
	ctx context.Context,
	req *service.GetExecutionRequest,
) (*service.GetExecutionResponse, error) {

	_, err := s.execStore.GetExecution(ctx, req.ProjectId, req.ExecutionId)
	if err != nil {
		return nil, err
	}

	return &service.GetExecutionResponse{
		//ExecutionId: exec.ID,
		//ProjectId:   exec.ProjectID,
		//WorkflowId:  exec.WorkflowID,
		//State:       exec.State,
	}, nil
}

func (s *WorkflowServer) StartWorkflow(ctx context.Context, req *service.StartWorkflowRequest) (*service.StartWorkflowResponse, error) {
	workflowID := fmt.Sprintf(
		"%s:%s:%s",
		req.ProjectId,
		req.WorkflowId,
		req.ClientRequestId,
	)

	// Convert structpb.Value map to interface{} map
	inputs := make(map[string]interface{}, len(req.Inputs))
	for k, v := range req.Inputs {
		if v != nil {
			inputs[k] = v.AsInterface()
		}
	}

	exec := &execution.Execution{
		ID: uuid.NewString(),

		ProjectID:  req.ProjectId,
		WorkflowID: req.WorkflowId,

		ClientRequestID: req.ClientRequestId,

		TemporalWorkflowID: workflowID,
		State:              "PENDING",
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

	def, err := registry.Get(req.ProjectId, req.WorkflowId)
	tc, err := temporal.GetClientForProject(req.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}
	g := dag.Build(def.Def)

	workflowOptions := client.StartWorkflowOptions{
		ID:        req.ProjectId + "-" + req.WorkflowId,
		TaskQueue: "workflow-task-queue",
	}

	we, err := tc.Client.ExecuteWorkflow(ctx, workflowOptions, temporal.WorkflowExecution, g, inputs)
	if err != nil {
		return nil, err
	}

	if err := s.execStore.MarkRunning(ctx, exec.ID, we.GetRunID()); err != nil {
		return nil, err
	}

	return &service.StartWorkflowResponse{
		ExecutionId: we.GetID(),
		State:       "RUNNING",
	}, nil
}

// func (s *WorkflowServer) GetExecution(ctx context.Context, req *service.GetExecutionRequest) (*service.GetExecutionResponse, error) {
// 	info, err := temporal.GetExecution(
// 		ctx,
// 		req.ProjectId,
// 		req.ExecutionId,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &service.GetExecutionResponse{
// 		State: service.ExecutionState(
// 			service.ExecutionState_value[string(info.State)],
// 		),
// 		Error: info.Error,
// 	}, nil
// }

func structpbToInterface(v *structpb.Value) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	return v.AsInterface(), nil
}
