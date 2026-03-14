CREATE TABLE IF NOT EXISTS challenge_writeups (
    id BIGSERIAL PRIMARY KEY,
    challenge_id BIGINT NOT NULL UNIQUE REFERENCES challenges(id) ON DELETE CASCADE,
    title VARCHAR(256) NOT NULL,
    content TEXT NOT NULL,
    visibility VARCHAR(16) NOT NULL DEFAULT 'private',
    release_at TIMESTAMPTZ DEFAULT NULL,
    created_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_challenge_writeups_visibility_release
    ON challenge_writeups(visibility, release_at);
