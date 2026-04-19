ALTER TABLE contest_awd_services
    DROP COLUMN IF EXISTS awd_checker_last_preview_result;

ALTER TABLE contest_awd_services
    DROP COLUMN IF EXISTS awd_checker_last_preview_at;

ALTER TABLE contest_awd_services
    DROP COLUMN IF EXISTS awd_checker_validation_state;
