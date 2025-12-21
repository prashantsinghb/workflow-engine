package validation

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
)

type Validator interface {
	Validate(ctx context.Context, req *Request) error
}

type Request struct {
	ProjectID  string
	Definition *api.Definition
	Modules    *registry.ModuleRegistry
}
