ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_validation_state VARCHAR(24) NOT NULL DEFAULT 'pending';

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_last_preview_at TIMESTAMP NULL;

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_last_preview_result TEXT NOT NULL DEFAULT '';
