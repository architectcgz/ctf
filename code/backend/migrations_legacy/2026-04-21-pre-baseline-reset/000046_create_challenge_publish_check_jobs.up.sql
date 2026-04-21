CREATE TABLE challenge_publish_check_jobs (
    id BIGSERIAL PRIMARY KEY,
    challenge_id BIGINT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    requested_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(32) NOT NULL,
    request_source VARCHAR(32) NOT NULL DEFAULT 'admin_publish',
    result_json TEXT NOT NULL DEFAULT '',
    failure_summary VARCHAR(512) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cp_jobs_challenge_created
    ON challenge_publish_check_jobs (challenge_id, created_at DESC);

CREATE INDEX idx_cp_jobs_status_created
    ON challenge_publish_check_jobs (status, created_at ASC);

CREATE UNIQUE INDEX idx_cp_jobs_challenge_active
    ON challenge_publish_check_jobs (challenge_id)
    WHERE status IN ('pending', 'running');
