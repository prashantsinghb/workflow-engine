CREATE TABLE module_http_specs (
  module_id UUID PRIMARY KEY
    REFERENCES modules(id) ON DELETE CASCADE,

  -- Request
  method TEXT NOT NULL
    CHECK (method IN ('GET','POST','PUT','PATCH','DELETE','HEAD','OPTIONS')),
  url TEXT NOT NULL,

  headers JSONB DEFAULT '{}'::jsonb,
  query_params JSONB DEFAULT '{}'::jsonb,
  body_template JSONB,

  -- Networking
  timeout_ms INTEGER NOT NULL DEFAULT 30000,

  -- Retry policy
  retry_count INTEGER NOT NULL DEFAULT 3,
  retry_backoff_ms INTEGER NOT NULL DEFAULT 500,
  retry_on_status JSONB DEFAULT '[500,502,503,504]'::jsonb,

  -- Auth
  auth_type TEXT NOT NULL DEFAULT 'none'
    CHECK (auth_type IN ('none','bearer','basic','api_key','oauth2')),
  auth_config JSONB,

  -- Observability
  success_codes JSONB DEFAULT '[200,201,202,204]'::jsonb,

  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
