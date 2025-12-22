package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
)

type nodeStore struct {
	db *sql.DB
}

func (s *nodeStore) Upsert(ctx context.Context, n *execution.ExecutionNode) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	input, _ := json.Marshal(n.Input)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO execution_nodes (
			id, execution_id, node_id,
			executor_type, status,
			attempt, max_attempts,
			input
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (execution_id, node_id)
		DO UPDATE SET
			executor_type = EXCLUDED.executor_type,
			status = EXCLUDED.status,
			attempt = EXCLUDED.attempt,
			max_attempts = EXCLUDED.max_attempts,
			input = EXCLUDED.input
	`,
		n.ID,
		n.ExecutionID,
		n.NodeID,
		n.ExecutorType,
		n.Status,
		n.Attempt,
		n.MaxAttempts,
		input,
	)
	return err
}

func (s *nodeStore) MarkRunning(ctx context.Context, executionID uuid.UUID, nodeID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE execution_nodes
		SET
			status = $3,
			started_at = now()
		WHERE execution_id = $1 AND node_id = $2
	`, executionID, nodeID, execution.NodeRunning)
	return err
}

func (s *nodeStore) MarkSucceeded(
	ctx context.Context,
	executionID uuid.UUID,
	nodeID string,
	output map[string]any,
) error {

	out, _ := json.Marshal(output)
	now := time.Now()

	_, err := s.db.ExecContext(ctx, `
		UPDATE execution_nodes
		SET
			status = $3,
			output = $4,
			completed_at = $5,
			duration_ms = EXTRACT(EPOCH FROM ($5 - started_at)) * 1000
		WHERE execution_id = $1 AND node_id = $2
	`, executionID, nodeID, execution.NodeSucceeded, out, now)

	return err
}

func (s *nodeStore) MarkFailed(
	ctx context.Context,
	executionID uuid.UUID,
	nodeID string,
	errPayload map[string]any,
) error {

	errJSON, _ := json.Marshal(errPayload)

	_, err := s.db.ExecContext(ctx, `
		UPDATE execution_nodes
		SET
			status = $3,
			error = $4,
			completed_at = now()
		WHERE execution_id = $1 AND node_id = $2
	`, executionID, nodeID, execution.NodeFailed, errJSON)

	return err
}

func (s *nodeStore) IncrementAttempt(
	ctx context.Context,
	executionID uuid.UUID,
	nodeID string,
) error {

	_, err := s.db.ExecContext(ctx, `
		UPDATE execution_nodes
		SET
			attempt = attempt + 1,
			status = $3
		WHERE execution_id = $1 AND node_id = $2
	`, executionID, nodeID, execution.NodeRetrying)

	return err
}

func (s *nodeStore) ListByExecution(
	ctx context.Context,
	executionID uuid.UUID,
) ([]execution.ExecutionNode, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id, execution_id, node_id,
			executor_type, status,
			attempt, max_attempts,
			input, output, error,
			started_at, completed_at, duration_ms
		FROM execution_nodes
		WHERE execution_id = $1
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []execution.ExecutionNode
	for rows.Next() {
		var n execution.ExecutionNode
		var input, output, errJSON []byte
		var started, completed sql.NullTime
		var duration sql.NullInt64

		if err := rows.Scan(
			&n.ID,
			&n.ExecutionID,
			&n.NodeID,
			&n.ExecutorType,
			&n.Status,
			&n.Attempt,
			&n.MaxAttempts,
			&input,
			&output,
			&errJSON,
			&started,
			&completed,
			&duration,
		); err != nil {
			return nil, err
		}

		_ = json.Unmarshal(input, &n.Input)
		_ = json.Unmarshal(output, &n.Output)
		_ = json.Unmarshal(errJSON, &n.Error)

		if started.Valid {
			n.StartedAt = &started.Time
		}
		if completed.Valid {
			n.CompletedAt = &completed.Time
		}
		if duration.Valid {
			n.DurationMs = &duration.Int64
		}

		nodes = append(nodes, n)
	}

	return nodes, nil
}

func (s *nodeStore) ListRunnable(
	ctx context.Context,
	executionID uuid.UUID,
) ([]execution.ExecutionNode, error) {
	// List nodes that are in PENDING or RETRYING status
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id, execution_id, node_id,
			executor_type, status,
			attempt, max_attempts,
			input, output, error,
			started_at, completed_at, duration_ms
		FROM execution_nodes
		WHERE execution_id = $1
			AND (status = $2 OR status = $3)
	`, executionID, execution.NodePending, execution.NodeRetrying)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []execution.ExecutionNode
	for rows.Next() {
		var n execution.ExecutionNode
		var input, output, errJSON []byte
		var started, completed sql.NullTime
		var duration sql.NullInt64

		if err := rows.Scan(
			&n.ID,
			&n.ExecutionID,
			&n.NodeID,
			&n.ExecutorType,
			&n.Status,
			&n.Attempt,
			&n.MaxAttempts,
			&input,
			&output,
			&errJSON,
			&started,
			&completed,
			&duration,
		); err != nil {
			return nil, err
		}

		_ = json.Unmarshal(input, &n.Input)
		_ = json.Unmarshal(output, &n.Output)
		_ = json.Unmarshal(errJSON, &n.Error)

		if started.Valid {
			n.StartedAt = &started.Time
		}
		if completed.Valid {
			n.CompletedAt = &completed.Time
		}
		if duration.Valid {
			n.DurationMs = &duration.Int64
		}

		nodes = append(nodes, n)
	}

	return nodes, nil
}
