ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_type VARCHAR(32) NOT NULL DEFAULT '';

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_config TEXT NOT NULL DEFAULT '{}';

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_sla_score INTEGER NOT NULL DEFAULT 0;

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_defense_score INTEGER NOT NULL DEFAULT 0;
