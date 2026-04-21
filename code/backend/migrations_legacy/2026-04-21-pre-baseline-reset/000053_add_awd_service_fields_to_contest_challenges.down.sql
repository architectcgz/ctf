ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_defense_score;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_sla_score;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_config;

ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS awd_checker_type;
