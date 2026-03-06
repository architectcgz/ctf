ALTER TABLE submissions
    ADD COLUMN IF NOT EXISTS contest_id BIGINT,
    ADD COLUMN IF NOT EXISTS team_id BIGINT,
    ADD COLUMN IF NOT EXISTS score INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_submissions_contest_id
    ON submissions(contest_id)
    WHERE contest_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_user_challenge_correct
    ON submissions(contest_id, user_id, challenge_id)
    WHERE is_correct = TRUE AND contest_id IS NOT NULL;
