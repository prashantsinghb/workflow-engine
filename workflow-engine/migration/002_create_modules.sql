CREATE TABLE modules (
  id UUID PRIMARY KEY,
  project_id TEXT NOT NULL, -- empty = global
  name TEXT NOT NULL,
  version TEXT NOT NULL,

  runtime TEXT NOT NULL, -- http | docker | internal

  inputs JSONB,
  outputs JSONB,

  created_at TIMESTAMPTZ DEFAULT now(),

  UNIQUE (project_id, name, version)
);