CREATE TABLE runtime_nodes (
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

ALTER TABLE instances
    ADD COLUMN node_id BIGINT REFERENCES runtime_nodes(id);

CREATE INDEX idx_instances_node_id ON instances(node_id);
