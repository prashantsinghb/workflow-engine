package postgres

import (
	"database/sql"

	"github.com/prashantsinghb/workflow-engine/pkg/execution"
)

type Store struct {
	executions execution.ExecutionStore
	nodes      execution.NodeStore
	events     execution.EventStore
}

func New(db *sql.DB) *Store {
	return &Store{
		executions: &executionStore{db: db},
		nodes:      &nodeStore{db: db},
		events:     &eventStore{db: db},
	}
}

func (s *Store) Executions() execution.ExecutionStore {
	return s.executions
}

func (s *Store) Nodes() execution.NodeStore {
	return s.nodes
}

func (s *Store) Events() execution.EventStore {
	return s.events
}
