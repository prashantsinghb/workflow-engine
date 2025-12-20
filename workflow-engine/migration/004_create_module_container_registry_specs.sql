CREATE TABLE module_container_registry_specs (
  module_id UUID PRIMARY KEY REFERENCES modules(id) ON DELETE CASCADE,

  image TEXT NOT NULL,
  command TEXT[],
  env JSONB,

  cpu TEXT,
  memory TEXT
);
