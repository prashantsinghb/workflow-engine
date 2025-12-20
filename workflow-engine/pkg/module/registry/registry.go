package registry

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/module/api"
)

type ModuleRegistry struct {
	store *PostgresRegistry
}

func NewModuleRegistry(store *PostgresRegistry) *ModuleRegistry {
	return &ModuleRegistry{store: store}
}

// GetStore returns the underlying PostgresRegistry for direct access
func (r *ModuleRegistry) GetStore() *PostgresRegistry {
	return r.store
}

// Register a module
func (r *ModuleRegistry) Register(ctx context.Context, m *api.Module) (string, error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.Version == "" {
		m.Version = "v1"
	}
	err := r.store.Insert(ctx, m)
	return m.ID, err
}

// Lookup module
func (r *ModuleRegistry) GetModule(ctx context.Context, projectID, name, version string) (*api.Module, error) {
	m, err := r.store.Get(ctx, projectID, name, version)
	if err != nil {
		return nil, fmt.Errorf("module not found: %w", err)
	}
	return m, nil
}

// List modules
func (r *ModuleRegistry) ListModules(ctx context.Context, projectID string) ([]*api.Module, error) {
	return r.store.List(ctx, projectID)
}
