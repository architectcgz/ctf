CREATE TABLE IF NOT EXISTS contest_challenges (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    points INTEGER NOT NULL DEFAULT 0,
    "order" INTEGER NOT NULL DEFAULT 0,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_contest_challenges_contest_id ON contest_challenges(contest_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_contest_challenges_active_unique
    ON contest_challenges(contest_id, challenge_id)
    WHERE deleted_at IS NULL;
