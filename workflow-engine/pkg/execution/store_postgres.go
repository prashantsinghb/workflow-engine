package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateExecution(ctx context.Context, exec *Execution) error {
	if exec.ID == "" {
		exec.ID = uuid.NewString()
	}

	inputs, _ := json.Marshal(exec.Inputs)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO executions (
			id, project_id, workflow_id, client_request_id,
			temporal_workflow_id,
			state, inputs
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (project_id, workflow_id, client_request_id)
		DO NOTHING
	`,
		exec.ID,
		exec.ProjectID,
		exec.WorkflowID,
		exec.ClientRequestID,
		exec.TemporalWorkflowID,
		exec.State,
		inputs,
	)

	return err
}

func (s *PostgresStore) GetExecution(ctx context.Context, projectID, executionID string) (*Execution, error) {
	// Try to parse as UUID first
	var row *sql.Row
	if _, err := uuid.Parse(executionID); err == nil {
		// Valid UUID - lookup by id
		row = s.db.QueryRowContext(ctx, `
			SELECT id, project_id, workflow_id,
			       temporal_workflow_id, temporal_run_id,
			       state, error, inputs, outputs,
			       started_at, completed_at, created_at, updated_at
			FROM executions
			WHERE project_id = $1 AND id = $2
		`, projectID, executionID)
	} else {
		// Not a valid UUID - try lookup by temporal_workflow_id
		row = s.db.QueryRowContext(ctx, `
			SELECT id, project_id, workflow_id,
			       temporal_workflow_id, temporal_run_id,
			       state, error, inputs, outputs,
			       started_at, completed_at, created_at, updated_at
			FROM executions
			WHERE project_id = $1 AND temporal_workflow_id = $2
		`, projectID, executionID)
	}

	var exec Execution
	var inputs, outputs []byte
	var errorStr sql.NullString
	var temporalRunID sql.NullString

	err := row.Scan(
		&exec.ID,
		&exec.ProjectID,
		&exec.WorkflowID,
		&exec.TemporalWorkflowID,
		&temporalRunID,
		&exec.State,
		&errorStr,
		&inputs,
		&outputs,
		&exec.StartedAt,
		&exec.CompletedAt,
		&exec.CreatedAt,
		&exec.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution not found: %s", executionID)
		}
		return nil, err
	}

	if temporalRunID.Valid {
		exec.TemporalRunID = temporalRunID.String
	}
	if errorStr.Valid {
		exec.Error = errorStr.String
	}

	_ = json.Unmarshal(inputs, &exec.Inputs)
	_ = json.Unmarshal(outputs, &exec.Outputs)

	return &exec, nil
}

func (s *PostgresStore) GetByIdempotencyKey(
	ctx context.Context,
	projectID string,
	workflowID string,
	clientRequestID string,
) (*Execution, error) {

	query := `
		SELECT
			id,
			project_id,
			workflow_id,
			client_request_id,
			temporal_workflow_id,
			temporal_run_id,
			state,
			error,
			inputs,
			outputs,
			started_at,
			completed_at,
			created_at,
			updated_at
		FROM executions
		WHERE project_id = $1
		  AND workflow_id = $2
		  AND client_request_id = $3
	`

	row := s.db.QueryRowContext(
		ctx,
		query,
		projectID,
		workflowID,
		clientRequestID,
	)

	var e Execution
	var errorStr sql.NullString
	var temporalRunID sql.NullString
	var inputs, outputs []byte

	if err := row.Scan(
		&e.ID,
		&e.ProjectID,
		&e.WorkflowID,
		&e.ClientRequestID,
		&e.TemporalWorkflowID,
		&temporalRunID,
		&e.State,
		&errorStr,
		&inputs,
		&outputs,
		&e.StartedAt,
		&e.CompletedAt,
		&e.CreatedAt,
		&e.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if temporalRunID.Valid {
		e.TemporalRunID = temporalRunID.String
	}
	if errorStr.Valid {
		e.Error = errorStr.String
	}
	_ = json.Unmarshal(inputs, &e.Inputs)
	_ = json.Unmarshal(outputs, &e.Outputs)

	return &e, nil
}

func (s *PostgresStore) MarkRunning(ctx context.Context, executionID, runID string) error {
	query := `
		UPDATE executions
		SET
			state = 'RUNNING',
			temporal_run_id = $2,
			started_at = now(),
			updated_at = now()
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, executionID, runID)
	return err
}

func (s *PostgresStore) ListRunningExecutions(ctx context.Context) ([]*Execution, error) {
	query := `SELECT id, project_id, workflow_id, client_request_id, temporal_workflow_id, state FROM executions WHERE state = 'RUNNING'`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*Execution
	for rows.Next() {
		var e Execution
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.WorkflowID, &e.ClientRequestID, &e.TemporalWorkflowID, &e.State); err != nil {
			return nil, err
		}
		executions = append(executions, &e)
	}
	return executions, nil
}

func (s *PostgresStore) MarkCompleted(ctx context.Context, executionID string, outputs map[string]interface{}) error {
	outputsJSON, _ := json.Marshal(outputs)
	if len(outputsJSON) == 0 || string(outputsJSON) == "null" {
		outputsJSON = []byte("{}")
	}

	query := `UPDATE executions SET state=$1, outputs=$2, completed_at=now(), updated_at=now() WHERE id=$3`
	_, err := s.db.ExecContext(ctx, query, StateSucceeded, string(outputsJSON), executionID)
	return err
}

func (s *PostgresStore) MarkFailed(ctx context.Context, executionID string, errMsg string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE executions
		SET state = $1,
		    error = $2,
		    completed_at = $3,
		    updated_at = now()
		WHERE id = $4
	`,
		StateFailed,
		errMsg,
		now,
		executionID,
	)
	return err
}

func (s *PostgresStore) UpdateNodeOutputs(ctx context.Context, nodeID string, outputs map[string]interface{}) error {
	query := `UPDATE nodes SET outputs=$2, updated_at=now() WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, nodeID, outputs)
	return err
}

func (s *PostgresStore) ListExecutions(ctx context.Context, projectID, workflowID string) ([]*Execution, error) {
	query := `SELECT id, project_id, workflow_id, client_request_id, state, error FROM executions WHERE project_id = $1`
	args := []interface{}{projectID}

	if workflowID != "" {
		query += ` AND workflow_id = $2`
		args = append(args, workflowID)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Execution
	for rows.Next() {
		e := &Execution{}
		var errorStr sql.NullString
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.WorkflowID, &e.ClientRequestID, &e.State, &errorStr); err != nil {
			return nil, err
		}
		if errorStr.Valid {
			e.Error = errorStr.String
		}
		result = append(result, e)
	}

	return result, nil
}

func (s *PostgresStore) GetStats(ctx context.Context, projectID string) (*ExecutionStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_executions,
			COUNT(*) FILTER (WHERE state = 'RUNNING') as running_executions,
			COUNT(*) FILTER (WHERE state = 'SUCCEEDED' OR state = 'COMPLETED') as success_count,
			COUNT(*) FILTER (WHERE state = 'FAILED') as failed_count
		FROM executions
		WHERE project_id = $1
	`

	var stats ExecutionStats
	err := s.db.QueryRowContext(ctx, query, projectID).Scan(
		&stats.TotalExecutions,
		&stats.RunningExecutions,
		&stats.SuccessCount,
		&stats.FailedCount,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
