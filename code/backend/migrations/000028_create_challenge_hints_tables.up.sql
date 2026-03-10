CREATE TABLE IF NOT EXISTS challenge_hints (
    id BIGSERIAL PRIMARY KEY,
    challenge_id BIGINT NOT NULL,
    level INT NOT NULL,
    title VARCHAR(128) DEFAULT '',
    cost_points INT NOT NULL DEFAULT 0,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_challenge_hints_challenge_level
    ON challenge_hints(challenge_id, level);

CREATE INDEX IF NOT EXISTS idx_challenge_hints_challenge_id
    ON challenge_hints(challenge_id);

CREATE TABLE IF NOT EXISTS challenge_hint_unlocks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    challenge_hint_id BIGINT NOT NULL,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_challenge_hint_unlocks_user_hint
    ON challenge_hint_unlocks(user_id, challenge_hint_id);

CREATE INDEX IF NOT EXISTS idx_challenge_hint_unlocks_user_challenge
    ON challenge_hint_unlocks(user_id, challenge_id);
