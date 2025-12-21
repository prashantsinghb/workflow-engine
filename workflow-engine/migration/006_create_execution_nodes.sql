CREATE TABLE execution_nodes (
  id UUID PRIMARY KEY,

  execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
  node_id TEXT NOT NULL,

  executor_type TEXT NOT NULL,
  -- http | container | grpc | subworkflow

  status TEXT NOT NULL,
  -- PENDING | RUNNING | SUCCEEDED | FAILED | SKIPPED | RETRYING

  attempt INT NOT NULL DEFAULT 1,
  max_attempts INT NOT NULL,

  input  JSONB,
  output JSONB,
  error  JSONB,

  started_at   TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  duration_ms BIGINT,

  UNIQUE (execution_id, node_id)
);

CREATE INDEX idx_exec_nodes_execution ON execution_nodes(execution_id);
CREATE INDEX idx_exec_nodes_status ON execution_nodes(status);
