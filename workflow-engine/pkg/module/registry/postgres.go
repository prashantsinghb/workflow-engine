package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
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
	if len(inputsJSON) == 0 || string(inputsJSON) == "null" {
		inputsJSON = []byte("{}")
	}

	outputsJSON, _ := json.Marshal(m.Outputs)
	if len(outputsJSON) == 0 || string(outputsJSON) == "null" {
		outputsJSON = []byte("{}")
	}

	// Handle empty project_id for global modules
	projectID := m.ProjectID
	if projectID == "" {
		projectID = "" // Keep as empty string for global modules
	}

	runtimeConfigJSON, _ := json.Marshal(m.RuntimeConfig)
	if len(runtimeConfigJSON) == 0 || string(runtimeConfigJSON) == "null" {
		runtimeConfigJSON = []byte("{}")
	}

	query := `
	INSERT INTO modules (id, name, version, project_id, runtime, runtime_config, inputs, outputs, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	ON CONFLICT (project_id, name, version) DO UPDATE 
	SET runtime=$5, runtime_config=$6, inputs=$7, outputs=$8
	`

	now := time.Now()
	_, err := s.DB.ExecContext(ctx, query,
		m.ID, m.Name, m.Version, projectID, m.Runtime,
		string(runtimeConfigJSON), string(inputsJSON), string(outputsJSON), now,
	)
	return err
}

// Lookup module (project-specific then global)
func (s *PostgresRegistry) Get(ctx context.Context, projectID, name, version string) (*api.Module, error) {
	var query string
	var row *sql.Row

	if version != "" {
		// If version is specified, filter by it directly
		query = `
		SELECT id, name, version, project_id, runtime, runtime_config, inputs, outputs, created_at
		FROM modules
		WHERE name=$1 AND version=$2 AND (project_id=$3 OR project_id IS NULL OR project_id = '')
		ORDER BY CASE WHEN project_id=$3 THEN 0 ELSE 1 END
		LIMIT 1
		`
		row = s.DB.QueryRowContext(ctx, query, name, version, projectID)
	} else {
		// If version is not specified, get the latest by created_at
		query = `
		SELECT id, name, version, project_id, runtime, runtime_config, inputs, outputs, created_at
		FROM modules
		WHERE name=$1 AND (project_id=$2 OR project_id IS NULL OR project_id = '')
		ORDER BY CASE WHEN project_id=$2 THEN 0 ELSE 1 END, created_at DESC
		LIMIT 1
		`
		row = s.DB.QueryRowContext(ctx, query, name, projectID)
	}

	var m api.Module
	var runtimeConfigJSON, inputsJSON, outputsJSON sql.NullString
	if err := row.Scan(&m.ID, &m.Name, &m.Version, &m.ProjectID, &m.Runtime, &runtimeConfigJSON, &inputsJSON, &outputsJSON, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.UpdatedAt = m.CreatedAt // Use created_at as updated_at since column doesn't exist

	if runtimeConfigJSON.Valid && runtimeConfigJSON.String != "" && runtimeConfigJSON.String != "null" {
		if err := json.Unmarshal([]byte(runtimeConfigJSON.String), &m.RuntimeConfig); err != nil {
			m.RuntimeConfig = make(map[string]interface{})
		}
		if m.RuntimeConfig == nil {
			m.RuntimeConfig = make(map[string]interface{})
		}
	} else {
		m.RuntimeConfig = make(map[string]interface{})
	}
	if inputsJSON.Valid && inputsJSON.String != "" {
		json.Unmarshal([]byte(inputsJSON.String), &m.Inputs)
	} else {
		m.Inputs = make(map[string]interface{})
	}
	if outputsJSON.Valid && outputsJSON.String != "" {
		json.Unmarshal([]byte(outputsJSON.String), &m.Outputs)
	} else {
		m.Outputs = make(map[string]interface{})
	}

	return &m, nil
}

// List modules (global + project)
func (s *PostgresRegistry) List(ctx context.Context, projectID string) ([]*api.Module, error) {
	query := `
	SELECT id, name, version, project_id, runtime, runtime_config, inputs, outputs, created_at
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
		var runtimeConfigJSON, inputsJSON, outputsJSON sql.NullString
		if err := rows.Scan(&m.ID, &m.Name, &m.Version, &m.ProjectID, &m.Runtime, &runtimeConfigJSON, &inputsJSON, &outputsJSON, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.UpdatedAt = m.CreatedAt // Use created_at as updated_at since column doesn't exist
		if runtimeConfigJSON.Valid && runtimeConfigJSON.String != "" && runtimeConfigJSON.String != "null" {
			if err := json.Unmarshal([]byte(runtimeConfigJSON.String), &m.RuntimeConfig); err != nil {
				m.RuntimeConfig = make(map[string]interface{})
			}
			if m.RuntimeConfig == nil {
				m.RuntimeConfig = make(map[string]interface{})
			}
		} else {
			m.RuntimeConfig = make(map[string]interface{})
		}
		if inputsJSON.Valid && inputsJSON.String != "" {
			json.Unmarshal([]byte(inputsJSON.String), &m.Inputs)
		} else {
			m.Inputs = make(map[string]interface{})
		}
		if outputsJSON.Valid && outputsJSON.String != "" {
			json.Unmarshal([]byte(outputsJSON.String), &m.Outputs)
		} else {
			m.Outputs = make(map[string]interface{})
		}
		modules = append(modules, &m)
	}
	return modules, nil
}

// InsertHttpSpec inserts HTTP spec for a module
func (s *PostgresRegistry) InsertHttpSpec(ctx context.Context, moduleID string, spec *service.HttpModuleSpec) error {
	headersJSON, _ := json.Marshal(spec.Headers)
	if len(headersJSON) == 0 || string(headersJSON) == "null" {
		headersJSON = []byte("{}")
	}

	queryParamsJSON, _ := json.Marshal(spec.QueryParams)
	if len(queryParamsJSON) == 0 || string(queryParamsJSON) == "null" {
		queryParamsJSON = []byte("{}")
	}

	var bodyTemplateJSON []byte
	if spec.BodyTemplate != nil {
		bodyTemplateJSON, _ = json.Marshal(spec.BodyTemplate)
		if len(bodyTemplateJSON) == 0 || string(bodyTemplateJSON) == "null" {
			bodyTemplateJSON = []byte("{}")
		}
	} else {
		bodyTemplateJSON = []byte("{}")
	}

	// Extract auth_type and auth_config from HttpAuth
	authType := "none"
	var authConfigJSON string
	if spec.Auth != nil && spec.Auth.Type != nil {
		authConfig := make(map[string]interface{})
		switch v := spec.Auth.Type.(type) {
		case *service.HttpAuth_Bearer:
			if v.Bearer != nil {
				authType = "bearer"
				authConfig["token"] = v.Bearer.Token
			}
		case *service.HttpAuth_ApiKey:
			if v.ApiKey != nil {
				authType = "api_key"
				authConfig["header"] = v.ApiKey.Header
				authConfig["value"] = v.ApiKey.Value
			}
		case *service.HttpAuth_Oauth2:
			if v.Oauth2 != nil {
				authType = "oauth2"
				authConfig["token_url"] = v.Oauth2.TokenUrl
				authConfig["client_id"] = v.Oauth2.ClientId
				authConfig["client_secret"] = v.Oauth2.ClientSecret
			}
		}
		if len(authConfig) > 0 {
			authConfigJSONBytes, _ := json.Marshal(authConfig)
			authConfigJSON = string(authConfigJSONBytes)
		}
	}

	query := `
	INSERT INTO module_http_specs (module_id, method, url, headers, query_params, body_template, timeout_ms, retry_count, auth_type, auth_config)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (module_id) DO UPDATE 
	SET method=$2, url=$3, headers=$4, query_params=$5, body_template=$6, timeout_ms=$7, retry_count=$8, auth_type=$9, auth_config=$10
	`

	timeoutMs := int32(30000)
	if spec.TimeoutMs > 0 {
		timeoutMs = spec.TimeoutMs
	}
	retryCount := int32(3)
	if spec.RetryCount > 0 {
		retryCount = spec.RetryCount
	}

	// Handle NULL for optional JSON fields - only set to NULL if truly empty, otherwise use "{}"
	var bodyTemplateSQL interface{}
	if spec.BodyTemplate == nil {
		bodyTemplateSQL = nil
	} else {
		bodyTemplateSQL = string(bodyTemplateJSON)
	}

	var authConfigSQL interface{}
	if authConfigJSON == "" {
		authConfigSQL = nil
	} else {
		authConfigSQL = authConfigJSON
	}

	_, err := s.DB.ExecContext(ctx, query,
		moduleID, spec.Method, spec.Url,
		string(headersJSON), string(queryParamsJSON), bodyTemplateSQL,
		timeoutMs, retryCount, authType, authConfigSQL,
	)
	return err
}

// GetHttpSpec retrieves HTTP spec for a module
func (s *PostgresRegistry) GetHttpSpec(ctx context.Context, moduleID string) (*service.HttpModuleSpec, error) {
	query := `
	SELECT method, url, headers, query_params, body_template, timeout_ms, retry_count, auth_type, auth_config
	FROM module_http_specs
	WHERE module_id=$1
	`

	row := s.DB.QueryRowContext(ctx, query, moduleID)
	var spec service.HttpModuleSpec
	var headersJSON, queryParamsJSON, bodyTemplateJSON, authType string
	var authConfigJSON sql.NullString
	if err := row.Scan(&spec.Method, &spec.Url, &headersJSON, &queryParamsJSON, &bodyTemplateJSON, &spec.TimeoutMs, &spec.RetryCount, &authType, &authConfigJSON); err != nil {
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

	// Parse and set auth
	authConfigJSONStr := ""
	if authConfigJSON.Valid {
		authConfigJSONStr = authConfigJSON.String
	}
	if authType != "" && authType != "none" && authConfigJSONStr != "" && authConfigJSONStr != "null" {
		var authConfig map[string]interface{}
		if err := json.Unmarshal([]byte(authConfigJSONStr), &authConfig); err == nil {
			auth := &service.HttpAuth{}
			switch authType {
			case "bearer":
				if token, ok := authConfig["token"].(string); ok {
					auth.Type = &service.HttpAuth_Bearer{
						Bearer: &service.BearerAuth{Token: token},
					}
				}
			case "api_key":
				if header, ok := authConfig["header"].(string); ok {
					if value, ok := authConfig["value"].(string); ok {
						auth.Type = &service.HttpAuth_ApiKey{
							ApiKey: &service.ApiKeyAuth{Header: header, Value: value},
						}
					}
				}
			case "oauth2":
				if tokenURL, ok := authConfig["token_url"].(string); ok {
					if clientID, ok := authConfig["client_id"].(string); ok {
						oauth2 := &service.OAuth2Auth{
							TokenUrl: tokenURL,
							ClientId: clientID,
						}
						if clientSecret, ok := authConfig["client_secret"].(string); ok {
							oauth2.ClientSecret = clientSecret
						}
						auth.Type = &service.HttpAuth_Oauth2{
							Oauth2: oauth2,
						}
					}
				}
			}
			spec.Auth = auth
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

func (r *ModuleRegistry) Resolve(
	ctx context.Context,
	projectID string,
	uses string,
) (*api.Module, error) {

	// uses format: name@version (version optional)
	name := uses
	version := ""

	if parts := strings.Split(uses, "@"); len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	}

	return r.GetModule(ctx, projectID, name, version)
}
