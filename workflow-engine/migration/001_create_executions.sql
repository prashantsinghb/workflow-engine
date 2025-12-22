CREATE TABLE executions (
  id UUID PRIMARY KEY,

  project_id  TEXT NOT NULL,
  workflow_id TEXT NOT NULL,

  client_request_id TEXT NOT NULL,

  state TEXT NOT NULL,
  error TEXT,

  inputs  JSONB,
  outputs JSONB,

  started_at   TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,

  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now(),

  UNIQUE (project_id, workflow_id, client_request_id)
);

CREATE INDEX idx_exec_project ON executions(project_id);
CREATE INDEX idx_exec_workflow ON executions(workflow_id);
CREATE INDEX idx_exec_state ON executions(state);
