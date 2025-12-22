package main

import (
	"context"
	"fmt"

	wfreg "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

// Service A: local step
type DNSWorkflow struct{}

func (DNSWorkflow) CreateDNSRecord(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("Service A: CreateDNSRecord called with", input)
	return map[string]interface{}{"record_id": "r123"}, nil
}

// Service B: local step
type IAMWorkflow struct{}

func (IAMWorkflow) AssignRoles(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("Service B: AssignRoles called with", input)
	return map[string]interface{}{"roles_assigned": true}, nil
}

// Initialize services: registers steps locally and in module registry
func initServices(ctx context.Context) error {
	// Service A
	if err := wfreg.RegisterAnnotated(ctx, DNSWorkflow{}, "service-a", stepRegistry, moduleRegistry); err != nil {
		return err
	}

	// Service B
	if err := wfreg.RegisterAnnotated(ctx, IAMWorkflow{}, "service-b", stepRegistry, moduleRegistry); err != nil {
		return err
	}

	return nil
}
