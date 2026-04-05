ALTER TABLE challenge_hints
    ADD COLUMN IF NOT EXISTS cost_points INT NOT NULL DEFAULT 0;

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
