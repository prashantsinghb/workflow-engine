package registry

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/parser"
)

type PostgresWorkflowStore struct {
	db *sql.DB
}

func NewPostgresWorkflowStore(db *sql.DB) *PostgresWorkflowStore {
	return &PostgresWorkflowStore{db: db}
}

func (s *PostgresWorkflowStore) Register(
	ctx context.Context,
	projectID string,
	wf *Workflow,
) (string, error) {

	id := uuid.NewString()
	def, err := parser.ParseWorkflow([]byte(wf.Yaml))
	if err != nil {
		return "", err
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO workflows (id, project_id, name, version, yaml)
		 VALUES ($1,$2,$3,$4,$5)`,
		id, projectID, wf.Name, wf.Version, wf.Yaml,
	)
	if err != nil {
		return "", err
	}

	wf.ID = id
	wf.Def = def
	return id, nil
}

func (s *PostgresWorkflowStore) Get(
	ctx context.Context,
	projectID string,
	workflowID string,
) (*Workflow, error) {

	var wf Workflow
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, version, yaml FROM workflows
		 WHERE id=$1 AND project_id=$2`,
		workflowID, projectID,
	).Scan(&wf.ID, &wf.Name, &wf.Version, &wf.Yaml)
	if err != nil {
		return nil, err
	}

	def, err := parser.ParseWorkflow([]byte(wf.Yaml))
	if err != nil {
		return nil, err
	}
	wf.Def = def
	return &wf, nil
}

func (s *PostgresWorkflowStore) List(
	ctx context.Context,
	projectID string,
) ([]*Workflow, error) {

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, version, yaml FROM workflows WHERE project_id=$1`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Workflow
	for rows.Next() {
		var wf Workflow
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.Version, &wf.Yaml); err != nil {
			return nil, err
		}
		out = append(out, &wf)
	}
	return out, nil
}

func (s *PostgresWorkflowStore) Count(
	ctx context.Context,
	projectID string,
) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM workflows WHERE project_id=$1`,
		projectID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresWorkflowStore) RegisterStep(ctx context.Context, def StepDefinition) error {
	metaJSON, _ := json.Marshal(def.Metadata)
	inputJSON, _ := json.Marshal(def.InputSchema)
	outputJSON, _ := json.Marshal(def.OutputSchema)

	query := `
    INSERT INTO workflow_steps (name, version, service, module_id, metadata, input_schema, output_schema)
    VALUES ($1,$2,$3,$4,$5,$6,$7)
    ON CONFLICT (name, version) DO UPDATE
    SET service=$3, module_id=$4, metadata=$5, input_schema=$6, output_schema=$7
    `
	_, err := r.db.ExecContext(ctx, query, def.Name, def.Version, def.Service, def.ModuleID, string(metaJSON), string(inputJSON), string(outputJSON))
	return err
}

func (r *PostgresWorkflowStore) GetStep(ctx context.Context, name, version string) (*StepDefinition, error) {
	if version == "" {
		version = "v1"
	}
	query := `SELECT name, version, service, module_id, metadata, input_schema, output_schema FROM workflow_steps WHERE name=$1 AND version=$2`
	row := r.db.QueryRowContext(ctx, query, name, version)

	var def StepDefinition
	var metaJSON, inputJSON, outputJSON string
	if err := row.Scan(&def.Name, &def.Version, &def.Service, &def.ModuleID, &metaJSON, &inputJSON, &outputJSON); err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metaJSON), &def.Metadata)
	json.Unmarshal([]byte(inputJSON), &def.InputSchema)
	json.Unmarshal([]byte(outputJSON), &def.OutputSchema)

	return &def, nil
}
