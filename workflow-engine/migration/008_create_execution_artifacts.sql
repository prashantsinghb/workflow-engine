CREATE TABLE execution_artifacts (
  id UUID PRIMARY KEY,

  execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
  node_id TEXT,

  artifact_type TEXT,
  -- log | report | file | image | metrics

  uri TEXT NOT NULL,
  metadata JSONB,

  created_at TIMESTAMPTZ DEFAULT now()
);
