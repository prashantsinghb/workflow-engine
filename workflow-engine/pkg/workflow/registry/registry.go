package registry

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/prashantsinghb/workflow-engine/pkg/module/api"
	moduleregistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
)

// StepDefinition holds step metadata
type StepDefinition struct {
	Name         string
	Version      string
	Service      string
	Executor     executor.Executor
	ModuleID     string
	InputSchema  map[string]string
	OutputSchema map[string]string
	Metadata     map[string]string
}

// StepRegistry interface
type StepRegistry interface {
	RegisterStep(ctx context.Context, def StepDefinition) error
	GetStep(name, version string) (*StepDefinition, error)
	ListSteps() []*StepDefinition
}

// LocalStepRegistry in-memory
type LocalStepRegistry struct {
	steps map[string]*StepDefinition
}

func NewLocalStepRegistry() *LocalStepRegistry {
	return &LocalStepRegistry{steps: make(map[string]*StepDefinition)}
}

func (r *LocalStepRegistry) RegisterStep(ctx context.Context, def StepDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("step name required")
	}
	if def.Version == "" {
		def.Version = "v1"
	}
	key := fmt.Sprintf("%s@%s", def.Name, def.Version)
	r.steps[key] = &def
	return nil
}

func (r *LocalStepRegistry) GetStep(name, version string) (*StepDefinition, error) {
	if version == "" {
		version = "v1"
	}
	key := fmt.Sprintf("%s@%s", name, version)
	s, ok := r.steps[key]
	if !ok {
		return nil, fmt.Errorf("step %s not found", key)
	}
	return s, nil
}

func (r *LocalStepRegistry) ListSteps() []*StepDefinition {
	list := make([]*StepDefinition, 0, len(r.steps))
	for _, s := range r.steps {
		list = append(list, s)
	}
	return list
}

// ------------------------
// Register a single function
// ------------------------
func RegisterFunctionStep(
	ctx context.Context,
	name, version, service string,
	fn func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error),
	stepRegistry StepRegistry,
	moduleRegistry *moduleregistry.ModuleRegistry,
	inputSchema, outputSchema map[string]string,
) error {

	if version == "" {
		version = "v1"
	}

	handler := executor.NewFuncExecutor(fn)

	// Step definition
	stepDef := StepDefinition{
		Name:         name,
		Version:      version,
		Service:      service,
		Executor:     handler,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
		Metadata:     map[string]string{"registered_by": service},
	}

	if err := stepRegistry.RegisterStep(ctx, stepDef); err != nil {
		return err
	}

	// Convert schema for module
	inputs := make(map[string]interface{}, len(inputSchema))
	for k, v := range inputSchema {
		inputs[k] = v
	}

	outputs := make(map[string]interface{}, len(outputSchema))
	for k, v := range outputSchema {
		outputs[k] = v
	}

	// Module registration
	mod := &api.Module{
		ID:      "", // ModuleRegistry will assign UUID
		Name:    name,
		Version: version,
		Runtime: "go",
		Inputs:  inputs,
		Outputs: outputs,
	}

	_, err := moduleRegistry.Register(ctx, mod)
	if err != nil {
		return err
	}

	// Bind ModuleID to StepDefinition
	stepDef.ModuleID = mod.ID
	return stepRegistry.RegisterStep(ctx, stepDef)
}

// ------------------------
// Annotated struct registration
// ------------------------
func RegisterAnnotated(
	ctx context.Context,
	target any,
	service string,
	stepRegistry StepRegistry,
	moduleRegistry *moduleregistry.ModuleRegistry,
) error {

	t := reflect.TypeOf(target)
	v := reflect.ValueOf(target)

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("RegisterAnnotated expects a struct, got %s", t.Kind())
	}

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		if err := validateMethod(m); err != nil {
			continue
		}

		stepName := toStepName(t.Name(), m.Name)

		handlerFn := func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			args := []reflect.Value{v, reflect.ValueOf(ctx), reflect.ValueOf(inputs)}
			out := m.Func.Call(args)

			switch len(out) {
			case 1: // error only
				if !out[0].IsNil() {
					return nil, out[0].Interface().(error)
				}
				return nil, nil
			case 2: // output + error
				if !out[1].IsNil() {
					return nil, out[1].Interface().(error)
				}
				if out[0].IsNil() {
					return nil, nil
				}
				return out[0].Interface().(map[string]interface{}), nil
			default:
				return nil, fmt.Errorf("invalid return signature for step %s", stepName)
			}
		}

		if err := RegisterFunctionStep(
			ctx,
			stepName,
			"v1",
			service,
			handlerFn,
			stepRegistry,
			moduleRegistry,
			map[string]string{},
			map[string]string{},
		); err != nil {
			return fmt.Errorf("failed to register step %s: %w", stepName, err)
		}
	}

	return nil
}

// ------------------------
// Helpers
// ------------------------
func validateMethod(m reflect.Method) error {
	if m.Type.NumIn() != 3 {
		return fmt.Errorf("method %s must have 2 params (context, input)", m.Name)
	}

	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !m.Type.In(1).Implements(ctxType) {
		return fmt.Errorf("first param must be context.Context")
	}

	switch m.Type.NumOut() {
	case 1:
		if !isError(m.Type.Out(0)) {
			return fmt.Errorf("single return must be error")
		}
	case 2:
		if !isError(m.Type.Out(1)) {
			return fmt.Errorf("second return must be error")
		}
	default:
		return fmt.Errorf("invalid return count")
	}

	return nil
}

func isError(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

func toStepName(structName, method string) string {
	return strings.ToLower(structName + "." + method)
}
