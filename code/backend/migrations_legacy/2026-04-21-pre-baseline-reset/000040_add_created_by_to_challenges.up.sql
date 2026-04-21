ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS created_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_challenges_created_by ON challenges(created_by);
