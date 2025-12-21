package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/prashantsinghb/workflow-engine/pkg/execution"
)

type eventStore struct {
	db *sql.DB
}

func (s *eventStore) Append(ctx context.Context, e *execution.ExecutionEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	payload, _ := json.Marshal(e.Payload)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO execution_events (
			id, execution_id, node_id,
			event_type, message, payload
		)
		VALUES ($1,$2,$3,$4,$5,$6)
	`,
		e.ID,
		e.ExecutionID,
		e.NodeID,
		e.EventType,
		e.Message,
		payload,
	)
	return err
}

func (s *eventStore) List(
	ctx context.Context,
	executionID uuid.UUID,
) ([]execution.ExecutionEvent, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id, execution_id, node_id,
			event_type, message, payload,
			created_at
		FROM execution_events
		WHERE execution_id = $1
		ORDER BY created_at ASC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []execution.ExecutionEvent
	for rows.Next() {
		var e execution.ExecutionEvent
		var payload []byte

		if err := rows.Scan(
			&e.ID,
			&e.ExecutionID,
			&e.NodeID,
			&e.EventType,
			&e.Message,
			&payload,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}

		_ = json.Unmarshal(payload, &e.Payload)
		events = append(events, e)
	}
	return events, nil
}
