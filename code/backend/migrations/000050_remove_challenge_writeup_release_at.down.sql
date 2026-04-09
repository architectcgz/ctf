ALTER TABLE challenge_writeups
    ADD COLUMN IF NOT EXISTS release_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_challenge_writeups_visibility_release
    ON challenge_writeups(visibility, release_at);
