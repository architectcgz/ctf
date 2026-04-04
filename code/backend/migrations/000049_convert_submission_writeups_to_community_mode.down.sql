DROP INDEX IF EXISTS idx_submission_writeups_recommended;
DROP INDEX IF EXISTS idx_submission_writeups_visibility_status;
DROP INDEX IF EXISTS idx_challenge_writeups_recommended;

ALTER TABLE submission_writeups
    ADD COLUMN IF NOT EXISTS review_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS reviewed_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS review_comment TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ DEFAULT NULL;

UPDATE submission_writeups
SET
    submission_status = CASE
        WHEN submission_status = 'published' THEN 'submitted'
        ELSE submission_status
    END,
    submitted_at = CASE
        WHEN published_at IS NOT NULL THEN published_at
        ELSE submitted_at
    END,
    review_status = COALESCE(NULLIF(review_status, ''), 'pending');

ALTER TABLE submission_writeups
    DROP COLUMN IF EXISTS visibility_status,
    DROP COLUMN IF EXISTS is_recommended,
    DROP COLUMN IF EXISTS recommended_at,
    DROP COLUMN IF EXISTS recommended_by,
    DROP COLUMN IF EXISTS published_at;

CREATE INDEX IF NOT EXISTS idx_submission_writeups_review_status
    ON submission_writeups(review_status, updated_at DESC);

ALTER TABLE challenge_writeups
    DROP COLUMN IF EXISTS is_recommended,
    DROP COLUMN IF EXISTS recommended_at,
    DROP COLUMN IF EXISTS recommended_by;
