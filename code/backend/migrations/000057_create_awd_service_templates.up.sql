CREATE TABLE IF NOT EXISTS awd_service_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    slug VARCHAR(128) NOT NULL,
    category VARCHAR(64) NOT NULL DEFAULT '',
    difficulty VARCHAR(32) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    service_type VARCHAR(32) NOT NULL,
    deployment_mode VARCHAR(32) NOT NULL,
    version VARCHAR(32) NOT NULL DEFAULT 'v1',
    status VARCHAR(24) NOT NULL DEFAULT 'draft',
    checker_type VARCHAR(32) NOT NULL DEFAULT '',
    checker_config TEXT NOT NULL DEFAULT '{}',
    flag_mode VARCHAR(32) NOT NULL DEFAULT '',
    flag_config TEXT NOT NULL DEFAULT '{}',
    defense_entry_mode VARCHAR(32) NOT NULL DEFAULT '',
    access_config TEXT NOT NULL DEFAULT '{}',
    runtime_config TEXT NOT NULL DEFAULT '{}',
    readiness_status VARCHAR(24) NOT NULL DEFAULT 'pending',
    readiness_report TEXT NOT NULL DEFAULT '',
    last_verified_at TIMESTAMP NULL,
    last_verified_by BIGINT NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_awd_service_templates_slug
    ON awd_service_templates(slug)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_awd_service_templates_status
    ON awd_service_templates(status, service_type);
