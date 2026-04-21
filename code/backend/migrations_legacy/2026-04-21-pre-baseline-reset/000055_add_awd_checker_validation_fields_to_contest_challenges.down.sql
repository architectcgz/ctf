ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_last_preview_result;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_last_preview_at;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_validation_state;
