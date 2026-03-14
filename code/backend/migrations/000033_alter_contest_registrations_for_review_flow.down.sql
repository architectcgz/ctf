DROP INDEX IF EXISTS idx_contest_reg_status;

ALTER TABLE contest_registrations
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    ALTER COLUMN status SET DEFAULT 'approved';
