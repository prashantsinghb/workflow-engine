CREATE TABLE module_http_specs (
  module_id UUID PRIMARY KEY REFERENCES modules(id) ON DELETE CASCADE,

  method TEXT NOT NULL,     -- GET, POST, PUT, DELETE
  url TEXT NOT NULL,

  headers JSONB,
  query_params JSONB,
  body_template JSONB,

  timeout_ms INTEGER DEFAULT 30000,
  retry_count INTEGER DEFAULT 3
);