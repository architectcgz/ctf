CREATE TABLE IF NOT EXISTS contest_awd_services (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    template_id BIGINT NULL,
    display_name VARCHAR(128) NOT NULL DEFAULT '',
    "order" INTEGER NOT NULL DEFAULT 0,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    score_config TEXT NOT NULL DEFAULT '{}',
    runtime_config TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_contest_awd_services_contest_challenge
    ON contest_awd_services(contest_id, challenge_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_contest_awd_services_contest_order
    ON contest_awd_services(contest_id, "order", id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_contest_awd_services_template
    ON contest_awd_services(template_id)
    WHERE deleted_at IS NULL AND template_id IS NOT NULL;
