package registry

import (
	"context"
	"database/sql"

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
