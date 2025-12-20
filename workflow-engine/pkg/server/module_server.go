package server

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/module/api"
	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"google.golang.org/protobuf/types/known/structpb"
)

type ModuleServer struct {
	service.UnimplementedModuleServiceServer
	Registry registry.ModuleRegistry
}

func (s *ModuleServer) RegisterModule(ctx context.Context, req *service.RegisterModuleRequest) (*service.RegisterModuleResponse, error) {
	inputs := make(map[string]interface{})
	for k, v := range req.Inputs {
		val, _ := structpbToInterface(v)
		inputs[k] = val
	}

	outputs := make(map[string]interface{})
	for k, v := range req.Outputs {
		val, _ := structpbToInterface(v)
		outputs[k] = val
	}

	// Treat "global" as empty string for global modules
	projectID := req.ProjectId
	if projectID == "global" {
		projectID = ""
	}

	module := &api.Module{
		Name:      req.Name,
		Version:   req.Version,
		ProjectID: projectID,
		Runtime:   req.Runtime,
		Inputs:    inputs,
		Outputs:   outputs,
	}

	id, err := s.Registry.Register(ctx, module)
	if err != nil {
		return nil, err
	}

	// Save HTTP spec if present
	if req.Runtime == "http" && req.GetHttp() != nil {
		postgresReg := s.Registry.GetStore()
		if err := postgresReg.InsertHttpSpec(ctx, id, req.GetHttp()); err != nil {
			return nil, err
		}
	}

	// Save container registry spec if present
	if req.Runtime == "docker" && req.GetContainerRegistry() != nil {
		postgresReg := s.Registry.GetStore()
		if err := postgresReg.InsertContainerSpec(ctx, id, req.GetContainerRegistry()); err != nil {
			return nil, err
		}
	}

	return &service.RegisterModuleResponse{ModuleId: id}, nil
}

func (s *ModuleServer) GetModule(
	ctx context.Context,
	req *service.GetModuleRequest,
) (*service.GetModuleResponse, error) {
	// Treat "global" as empty string for global modules
	projectID := req.ProjectId
	if projectID == "global" {
		projectID = ""
	}
	m, err := s.Registry.GetModule(ctx, projectID, req.Name, req.Version)
	if err != nil {
		return nil, err
	}

	inputs, _ := structpb.NewStruct(m.Inputs)
	outputs, _ := structpb.NewStruct(m.Outputs)

	serviceModule := &service.Module{
		Id:        m.ID,
		ProjectId: m.ProjectID,
		Name:      m.Name,
		Version:   m.Version,
		Runtime:   m.Runtime,
	}

	// Load spec from appropriate table
	postgresReg := s.Registry.GetStore()
	if m.Runtime == "http" {
		httpSpec, err := postgresReg.GetHttpSpec(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		if httpSpec != nil {
			serviceModule.Spec = &service.Module_Http{
				Http: httpSpec,
			}
		} else {
			serviceModule.Spec = &service.Module_Http{
				Http: &service.HttpModuleSpec{},
			}
		}
	} else if m.Runtime == "docker" {
		containerSpec, err := postgresReg.GetContainerSpec(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		if containerSpec != nil {
			serviceModule.Spec = &service.Module_ContainerRegistry{
				ContainerRegistry: containerSpec,
			}
		} else {
			serviceModule.Spec = &service.Module_ContainerRegistry{
				ContainerRegistry: &service.ContainerRegistryModuleSpec{},
			}
		}
	}

	if inputs != nil {
		serviceModule.Inputs = inputs.Fields
	}
	if outputs != nil {
		serviceModule.Outputs = outputs.Fields
	}

	return &service.GetModuleResponse{Module: serviceModule}, nil
}

func (s *ModuleServer) ListModules(
	ctx context.Context,
	req *service.ListModulesRequest,
) (*service.ListModulesResponse, error) {
	// Treat "global" as empty string for global modules
	projectID := req.ProjectId
	if projectID == "global" {
		projectID = ""
	}
	apiModules, err := s.Registry.ListModules(ctx, projectID)
	if err != nil {
		return nil, err
	}

	serviceModules := make([]*service.Module, 0, len(apiModules))
	for _, m := range apiModules {
		inputs, _ := structpb.NewStruct(m.Inputs)
		outputs, _ := structpb.NewStruct(m.Outputs)

		serviceModule := &service.Module{
			Id:        m.ID,
			ProjectId: m.ProjectID,
			Name:      m.Name,
			Version:   m.Version,
			Runtime:   m.Runtime,
		}

		// Load spec from appropriate table
		postgresReg := s.Registry.GetStore()
		if m.Runtime == "http" {
			httpSpec, err := postgresReg.GetHttpSpec(ctx, m.ID)
			if err == nil && httpSpec != nil {
				serviceModule.Spec = &service.Module_Http{
					Http: httpSpec,
				}
			} else {
				serviceModule.Spec = &service.Module_Http{
					Http: &service.HttpModuleSpec{},
				}
			}
		} else if m.Runtime == "docker" {
			containerSpec, err := postgresReg.GetContainerSpec(ctx, m.ID)
			if err == nil && containerSpec != nil {
				serviceModule.Spec = &service.Module_ContainerRegistry{
					ContainerRegistry: containerSpec,
				}
			} else {
				serviceModule.Spec = &service.Module_ContainerRegistry{
					ContainerRegistry: &service.ContainerRegistryModuleSpec{},
				}
			}
		}

		if inputs != nil {
			serviceModule.Inputs = inputs.Fields
		}
		if outputs != nil {
			serviceModule.Outputs = outputs.Fields
		}

		serviceModules = append(serviceModules, serviceModule)
	}

	return &service.ListModulesResponse{Modules: serviceModules}, nil
}
