CREATE TABLE IF NOT EXISTS submission_writeups (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id BIGINT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    contest_id BIGINT DEFAULT NULL REFERENCES contests(id) ON DELETE SET NULL,
    title VARCHAR(256) NOT NULL,
    content TEXT NOT NULL,
    submission_status VARCHAR(16) NOT NULL DEFAULT 'draft',
    review_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    submitted_at TIMESTAMPTZ DEFAULT NULL,
    reviewed_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ DEFAULT NULL,
    review_comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_submission_writeups_user_challenge
    ON submission_writeups(user_id, challenge_id);

CREATE INDEX IF NOT EXISTS idx_submission_writeups_review_status
    ON submission_writeups(review_status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_submission_writeups_user_updated_at
    ON submission_writeups(user_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_submission_writeups_challenge
    ON submission_writeups(challenge_id, updated_at DESC);
