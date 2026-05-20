CREATE TABLE network_allocations (
    subnet TEXT PRIMARY KEY,
    instance_id BIGINT NULL,
    network_key VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_network_allocations_instance_id ON network_allocations (instance_id);

CREATE UNIQUE INDEX uk_network_allocations_owner_key
    ON network_allocations (instance_id, network_key);
