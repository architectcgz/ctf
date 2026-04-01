DROP INDEX IF EXISTS idx_submissions_review_status;

ALTER TABLE submissions
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS review_comment,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS review_status;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS flag_regex;
