DROP INDEX IF EXISTS uk_submissions_contest_user_challenge_correct;
DROP INDEX IF EXISTS idx_submissions_contest_id;

ALTER TABLE submissions
    DROP COLUMN IF EXISTS score,
    DROP COLUMN IF EXISTS team_id,
    DROP COLUMN IF EXISTS contest_id;
