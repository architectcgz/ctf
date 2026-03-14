ALTER TABLE contest_registrations
    ALTER COLUMN status SET DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS reviewed_by BIGINT,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_contest_reg_status
    ON contest_registrations(contest_id, status);
