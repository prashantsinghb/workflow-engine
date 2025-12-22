package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
)

type executionStore struct {
	db *sql.DB
}

func (s *executionStore) Create(ctx context.Context, e *execution.Execution) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	inputs, _ := json.Marshal(e.Inputs)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO executions (
			id,
			project_id,
			workflow_id,
			client_request_id,
			state,
			inputs
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (project_id, workflow_id, client_request_id)
		DO NOTHING
	`,
		e.ID,
		e.ProjectID,
		e.WorkflowID,
		e.ClientRequestID,
		execution.ExecutionPending,
		inputs,
	)
	return err
}

func (s *executionStore) Get(
	ctx context.Context,
	projectID string,
	executionID uuid.UUID,
) (*execution.Execution, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT
			id, project_id, workflow_id,
			client_request_id,
			state, error,
			inputs, outputs,
			started_at, completed_at,
			created_at, updated_at
		FROM executions
		WHERE project_id = $1 AND id = $2
	`, projectID, executionID)

	return scanExecution(row)
}

func (s *executionStore) GetByIdempotencyKey(
	ctx context.Context,
	projectID, workflowID, clientRequestID string,
) (*execution.Execution, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT
			id, project_id, workflow_id,
			client_request_id,
			state, error,
			inputs, outputs,
			started_at, completed_at,
			created_at, updated_at
		FROM executions
		WHERE project_id=$1 AND workflow_id=$2 AND client_request_id=$3
	`, projectID, workflowID, clientRequestID)

	return scanExecution(row)
}

func (s *executionStore) MarkRunning(
	ctx context.Context,
	executionID uuid.UUID,
) error {

	_, err := s.db.ExecContext(ctx, `
		UPDATE executions
		SET
			state = $2,
			started_at = now(),
			updated_at = now()
		WHERE id = $1
	`,
		executionID,
		execution.ExecutionRunning,
	)
	return err
}

func (s *executionStore) MarkCompleted(
	ctx context.Context,
	executionID uuid.UUID,
	outputs map[string]any,
) error {

	out, _ := json.Marshal(outputs)

	_, err := s.db.ExecContext(ctx, `
		UPDATE executions
		SET
			state = $2,
			outputs = $3,
			completed_at = now(),
			updated_at = now()
		WHERE id = $1
	`,
		executionID,
		execution.ExecutionSucceeded,
		out,
	)
	return err
}

func (s *executionStore) MarkFailed(
	ctx context.Context,
	executionID uuid.UUID,
	errPayload map[string]any,
) error {

	errJSON, _ := json.Marshal(errPayload)

	_, err := s.db.ExecContext(ctx, `
		UPDATE executions
		SET
			state = $2,
			error = $3,
			completed_at = now(),
			updated_at = now()
		WHERE id = $1
	`,
		executionID,
		execution.ExecutionFailed,
		errJSON,
	)
	return err
}

func (s *executionStore) List(
	ctx context.Context,
	projectID, workflowID string,
) ([]*execution.Execution, error) {

	query := `
		SELECT
			id, project_id, workflow_id,
			client_request_id,
			state, error,
			inputs, outputs,
			started_at, completed_at,
			created_at, updated_at
		FROM executions
		WHERE project_id = $1
	`
	args := []any{projectID}

	if workflowID != "" {
		query += " AND workflow_id = $2"
		args = append(args, workflowID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*execution.Execution
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

func (s *executionStore) ListRunning(ctx context.Context) ([]*execution.Execution, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id, project_id, workflow_id,
			client_request_id,
			state,
			started_at,
			created_at, updated_at
		FROM executions
		WHERE state = $1
	`, execution.ExecutionRunning)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*execution.Execution
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

func (s *executionStore) GetStats(
	ctx context.Context,
	projectID string,
) (*execution.ExecutionStats, error) {

	var stats execution.ExecutionStats
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE state = 'RUNNING'),
			COUNT(*) FILTER (WHERE state = 'SUCCEEDED'),
			COUNT(*) FILTER (WHERE state = 'FAILED')
		FROM executions
		WHERE project_id = $1
	`, projectID).Scan(
		&stats.TotalExecutions,
		&stats.RunningExecutions,
		&stats.SuccessCount,
		&stats.FailedCount,
	)
	return &stats, err
}

func scanExecution(row interface {
	Scan(dest ...any) error
}) (*execution.Execution, error) {

	var e execution.Execution
	var inputs, outputs, errJSON []byte
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&e.ID,
		&e.ProjectID,
		&e.WorkflowID,
		&e.ClientRequestID,
		&e.Status,
		&errJSON,
		&inputs,
		&outputs,
		&startedAt,
		&completedAt,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		e.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		e.CompletedAt = &completedAt.Time
	}

	_ = json.Unmarshal(inputs, &e.Inputs)
	_ = json.Unmarshal(outputs, &e.Outputs)
	_ = json.Unmarshal(errJSON, &e.Error)

	return &e, nil
}
