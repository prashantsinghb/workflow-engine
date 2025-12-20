package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/module/api"
	"google.golang.org/protobuf/types/known/structpb"
)

type PostgresRegistry struct {
	DB *sql.DB
}

func NewPostgresRegistry(db *sql.DB) *PostgresRegistry {
	return &PostgresRegistry{DB: db}
}

func (s *PostgresRegistry) Insert(ctx context.Context, m *api.Module) error {
	inputsJSON, _ := json.Marshal(m.Inputs)
	outputsJSON, _ := json.Marshal(m.Outputs)

	// Handle empty project_id for global modules
	projectID := m.ProjectID
	if projectID == "" {
		projectID = "" // Keep as empty string for global modules
	}

	query := `
	INSERT INTO modules (id, name, version, project_id, runtime, inputs, outputs, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	ON CONFLICT (project_id, name, version) DO UPDATE 
	SET runtime=$5, inputs=$6, outputs=$7
	`

	now := time.Now()
	_, err := s.DB.ExecContext(ctx, query,
		m.ID, m.Name, m.Version, projectID, m.Runtime,
		string(inputsJSON), string(outputsJSON), now,
	)
	return err
}

// Lookup module (project-specific then global)
func (s *PostgresRegistry) Get(ctx context.Context, projectID, name, version string) (*api.Module, error) {
	query := `
	SELECT id, name, version, project_id, runtime, inputs, outputs, created_at
	FROM modules
	WHERE name=$1 AND (project_id=$2 OR project_id IS NULL OR project_id = '')
	ORDER BY CASE WHEN project_id=$2 THEN 0 ELSE 1 END, created_at DESC
	LIMIT 1
	`

	row := s.DB.QueryRowContext(ctx, query, name, projectID)
	var m api.Module
	var inputsJSON, outputsJSON string
	if err := row.Scan(&m.ID, &m.Name, &m.Version, &m.ProjectID, &m.Runtime, &inputsJSON, &outputsJSON, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.UpdatedAt = m.CreatedAt // Use created_at as updated_at since column doesn't exist

	json.Unmarshal([]byte(inputsJSON), &m.Inputs)
	json.Unmarshal([]byte(outputsJSON), &m.Outputs)

	if version != "" && m.Version != version {
		return nil, fmt.Errorf("module version not found")
	}

	return &m, nil
}

// List modules (global + project)
func (s *PostgresRegistry) List(ctx context.Context, projectID string) ([]*api.Module, error) {
	query := `
	SELECT id, name, version, project_id, runtime, inputs, outputs, created_at
	FROM modules
	WHERE project_id=$1 OR project_id IS NULL OR project_id = ''
	ORDER BY name, version
	`

	rows, err := s.DB.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*api.Module
	for rows.Next() {
		var m api.Module
		var inputsJSON, outputsJSON string
		if err := rows.Scan(&m.ID, &m.Name, &m.Version, &m.ProjectID, &m.Runtime, &inputsJSON, &outputsJSON, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.UpdatedAt = m.CreatedAt // Use created_at as updated_at since column doesn't exist
		json.Unmarshal([]byte(inputsJSON), &m.Inputs)
		json.Unmarshal([]byte(outputsJSON), &m.Outputs)
		modules = append(modules, &m)
	}
	return modules, nil
}

// InsertHttpSpec inserts HTTP spec for a module
func (s *PostgresRegistry) InsertHttpSpec(ctx context.Context, moduleID string, spec *service.HttpModuleSpec) error {
	headersJSON, _ := json.Marshal(spec.Headers)
	queryParamsJSON, _ := json.Marshal(spec.QueryParams)
	bodyTemplateJSON, _ := json.Marshal(spec.BodyTemplate)

	query := `
	INSERT INTO module_http_specs (module_id, method, url, headers, query_params, body_template, timeout_ms, retry_count)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (module_id) DO UPDATE 
	SET method=$2, url=$3, headers=$4, query_params=$5, body_template=$6, timeout_ms=$7, retry_count=$8
	`

	timeoutMs := int32(30000)
	if spec.TimeoutMs > 0 {
		timeoutMs = spec.TimeoutMs
	}
	retryCount := int32(3)
	if spec.RetryCount > 0 {
		retryCount = spec.RetryCount
	}

	_, err := s.DB.ExecContext(ctx, query,
		moduleID, spec.Method, spec.Url,
		string(headersJSON), string(queryParamsJSON), string(bodyTemplateJSON),
		timeoutMs, retryCount,
	)
	return err
}

// GetHttpSpec retrieves HTTP spec for a module
func (s *PostgresRegistry) GetHttpSpec(ctx context.Context, moduleID string) (*service.HttpModuleSpec, error) {
	query := `
	SELECT method, url, headers, query_params, body_template, timeout_ms, retry_count
	FROM module_http_specs
	WHERE module_id=$1
	`

	row := s.DB.QueryRowContext(ctx, query, moduleID)
	var spec service.HttpModuleSpec
	var headersJSON, queryParamsJSON, bodyTemplateJSON string
	if err := row.Scan(&spec.Method, &spec.Url, &headersJSON, &queryParamsJSON, &bodyTemplateJSON, &spec.TimeoutMs, &spec.RetryCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No spec found
		}
		return nil, err
	}

	// Parse JSON fields
	var headers map[string]string
	var queryParams map[string]string
	json.Unmarshal([]byte(headersJSON), &headers)
	json.Unmarshal([]byte(queryParamsJSON), &queryParams)
	spec.Headers = headers
	spec.QueryParams = queryParams

	// Parse body_template
	if bodyTemplateJSON != "" && bodyTemplateJSON != "null" {
		var bodyTemplate map[string]interface{}
		if err := json.Unmarshal([]byte(bodyTemplateJSON), &bodyTemplate); err == nil {
			spec.BodyTemplate, _ = structpb.NewStruct(bodyTemplate)
		}
	}

	return &spec, nil
}

// InsertContainerSpec inserts container registry spec for a module
func (s *PostgresRegistry) InsertContainerSpec(ctx context.Context, moduleID string, spec *service.ContainerRegistryModuleSpec) error {
	envJSON, _ := json.Marshal(spec.Env)

	query := `
	INSERT INTO module_container_registry_specs (module_id, image, command, env, cpu, memory)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (module_id) DO UPDATE 
	SET image=$2, command=$3, env=$4, cpu=$5, memory=$6
	`

	_, err := s.DB.ExecContext(ctx, query,
		moduleID, spec.Image, pq.Array(spec.Command), string(envJSON), spec.Cpu, spec.Memory,
	)
	return err
}

// GetContainerSpec retrieves container registry spec for a module
func (s *PostgresRegistry) GetContainerSpec(ctx context.Context, moduleID string) (*service.ContainerRegistryModuleSpec, error) {
	query := `
	SELECT image, command, env, cpu, memory
	FROM module_container_registry_specs
	WHERE module_id=$1
	`

	row := s.DB.QueryRowContext(ctx, query, moduleID)
	var spec service.ContainerRegistryModuleSpec
	var envJSON string
	var commandArray pq.StringArray
	if err := row.Scan(&spec.Image, &commandArray, &envJSON, &spec.Cpu, &spec.Memory); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No spec found
		}
		return nil, err
	}

	spec.Command = []string(commandArray)
	var env map[string]string
	json.Unmarshal([]byte(envJSON), &env)
	spec.Env = env

	return &spec, nil
}
