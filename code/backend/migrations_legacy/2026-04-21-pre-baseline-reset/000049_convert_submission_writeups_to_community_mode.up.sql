ALTER TABLE challenge_writeups
    ADD COLUMN IF NOT EXISTS is_recommended BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recommended_at TIMESTAMPTZ DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS recommended_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_challenge_writeups_recommended
    ON challenge_writeups(is_recommended, recommended_at DESC);

ALTER TABLE submission_writeups
    ADD COLUMN IF NOT EXISTS visibility_status VARCHAR(16) NOT NULL DEFAULT 'visible',
    ADD COLUMN IF NOT EXISTS is_recommended BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recommended_at TIMESTAMPTZ DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS recommended_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ DEFAULT NULL;

UPDATE submission_writeups
SET
    submission_status = CASE
        WHEN submission_status = 'submitted' THEN 'published'
        ELSE submission_status
    END,
    visibility_status = COALESCE(NULLIF(visibility_status, ''), 'visible'),
    is_recommended = COALESCE(is_recommended, FALSE),
    published_at = CASE
        WHEN published_at IS NOT NULL THEN published_at
        WHEN submitted_at IS NOT NULL THEN submitted_at
        WHEN submission_status = 'published' THEN updated_at
        ELSE NULL
    END;

DROP INDEX IF EXISTS idx_submission_writeups_review_status;

ALTER TABLE submission_writeups
    DROP COLUMN IF EXISTS review_status,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS review_comment,
    DROP COLUMN IF EXISTS submitted_at;

CREATE INDEX IF NOT EXISTS idx_submission_writeups_visibility_status
    ON submission_writeups(visibility_status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_submission_writeups_recommended
    ON submission_writeups(is_recommended, recommended_at DESC);
