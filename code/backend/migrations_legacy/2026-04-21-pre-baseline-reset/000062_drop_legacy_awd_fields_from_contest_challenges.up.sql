ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_last_preview_result;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_last_preview_at;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_validation_state;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_defense_score;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_sla_score;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_config;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_type;
