CREATE TABLE IF NOT EXISTS runtime_nodes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    endpoint VARCHAR(255) NOT NULL DEFAULT '',
    tls_identity VARCHAR(255) NOT NULL DEFAULT '',
    schedulable BOOLEAN NOT NULL DEFAULT TRUE,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    capacity_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE runtime_nodes
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET DEFAULT now();

UPDATE runtime_nodes
SET created_at = COALESCE(created_at, now()),
    updated_at = COALESCE(updated_at, now())
WHERE created_at IS NULL
   OR updated_at IS NULL;

ALTER TABLE runtime_nodes
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS node_id BIGINT REFERENCES runtime_nodes(id);

CREATE INDEX IF NOT EXISTS idx_instances_node_id ON instances(node_id);
