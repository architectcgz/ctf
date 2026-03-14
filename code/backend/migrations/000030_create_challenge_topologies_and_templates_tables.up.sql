CREATE TABLE IF NOT EXISTS environment_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    entry_node_key VARCHAR(64) NOT NULL,
    spec JSONB NOT NULL DEFAULT '{}'::jsonb,
    usage_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_environment_templates_name
    ON environment_templates(name);

CREATE TABLE IF NOT EXISTS challenge_topologies (
    id BIGSERIAL PRIMARY KEY,
    challenge_id BIGINT NOT NULL UNIQUE REFERENCES challenges(id) ON DELETE CASCADE,
    template_id BIGINT DEFAULT NULL REFERENCES environment_templates(id) ON DELETE SET NULL,
    entry_node_key VARCHAR(64) NOT NULL,
    spec JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_challenge_topologies_template_id
    ON challenge_topologies(template_id);
