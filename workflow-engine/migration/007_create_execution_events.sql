CREATE TABLE execution_events (
  id UUID PRIMARY KEY,

  execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
  node_id TEXT,

  event_type TEXT NOT NULL,
  -- EXECUTION_STARTED
  -- NODE_STARTED
  -- NODE_RETRY
  -- NODE_SUCCEEDED
  -- NODE_FAILED
  -- EXECUTION_PAUSED
  -- EXECUTION_RESUMED
  -- EXECUTION_CANCELLED

  message TEXT,
  payload JSONB,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exec_events_execution ON execution_events(execution_id);
CREATE INDEX idx_exec_events_node ON execution_events(node_id);
