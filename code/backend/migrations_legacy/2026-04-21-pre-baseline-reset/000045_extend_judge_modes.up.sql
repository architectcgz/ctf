ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS flag_regex TEXT NOT NULL DEFAULT '';

ALTER TABLE submissions
    ADD COLUMN IF NOT EXISTS review_status VARCHAR(32) NOT NULL DEFAULT 'not_required',
    ADD COLUMN IF NOT EXISTS reviewed_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS review_comment TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE submissions
SET updated_at = COALESCE(updated_at, submitted_at, NOW())
WHERE updated_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_submissions_review_status
    ON submissions(review_status, updated_at DESC);
