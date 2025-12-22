package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/module/api"
	moduleregistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
	wfRegistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

type StepRegistryServer struct {
	pb.UnimplementedStepRegistryServer
	stepRegistry   *wfRegistry.LocalStepRegistry
	moduleRegistry *moduleregistry.ModuleRegistry
}

func NewStepRegistryServer(stepRegistry *wfRegistry.LocalStepRegistry, moduleRegistry *moduleregistry.ModuleRegistry) *StepRegistryServer {
	return &StepRegistryServer{
		stepRegistry:   stepRegistry,
		moduleRegistry: moduleRegistry,
	}
}

func (s *StepRegistryServer) RegisterSteps(ctx context.Context, req *pb.RegisterStepsRequest) (*pb.RegisterStepsResponse, error) {
	for _, step := range req.Steps {
		stepDef := wfRegistry.StepDefinition{
			Name:         step.Name,
			Version:      step.Version,
			Service:      req.Service,
			Protocol:     step.Protocol,
			Endpoint:     step.Endpoint,
			InputSchema:  step.InputSchema,
			OutputSchema: step.OutputSchema,
			Metadata:     map[string]string{"registered_by": req.Service},
		}

		if err := s.stepRegistry.RegisterStep(ctx, stepDef); err != nil {
			return &pb.RegisterStepsResponse{Success: false, Message: err.Error()}, nil
		}

		// Convert input/output schemas from map[string]string to map[string]interface{}
		inputs := make(map[string]interface{}, len(step.InputSchema))
		for k, v := range step.InputSchema {
			inputs[k] = v
		}

		outputs := make(map[string]interface{}, len(step.OutputSchema))
		for k, v := range step.OutputSchema {
			outputs[k] = v
		}

		mod := &api.Module{
			ID:      uuid.New().String(),
			Name:    step.Name,
			Version: step.Version,
			Runtime: step.Protocol,
			Inputs:  inputs,
			Outputs: outputs,
			RuntimeConfig: map[string]interface{}{
				"endpoint": step.Endpoint,
				"protocol": step.Protocol,
				"service":  req.Service,
			},
		}

		if _, err := s.moduleRegistry.Register(ctx, mod); err != nil {
			return &pb.RegisterStepsResponse{Success: false, Message: err.Error()}, nil
		}
	}

	return &pb.RegisterStepsResponse{Success: true, Message: "steps registered successfully"}, nil
}
