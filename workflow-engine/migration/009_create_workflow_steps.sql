CREATE TABLE workflow_steps (
    id SERIAL PRIMARY KEY,          -- optional unique identifier
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    service TEXT NOT NULL,
    module_id UUID NOT NULL,
    metadata JSONB,
    input_schema JSONB,
    output_schema JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Optional: add an index on module_id if you query by it frequently
CREATE INDEX idx_workflow_steps_module_id ON workflow_steps(module_id);

-- Optional: unique constraint to prevent duplicate workflow step versions for the same name
ALTER TABLE workflow_steps
ADD CONSTRAINT uq_workflow_steps_name_version UNIQUE (name, version);
