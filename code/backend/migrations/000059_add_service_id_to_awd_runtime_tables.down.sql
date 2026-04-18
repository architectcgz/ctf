DROP INDEX IF EXISTS idx_awd_attack_round_service_success;
DROP INDEX IF EXISTS idx_awd_ts_round_team_service;
DROP INDEX IF EXISTS uk_awd_team_services;

CREATE UNIQUE INDEX IF NOT EXISTS uk_awd_team_services
    ON awd_team_services(round_id, team_id, challenge_id);

ALTER TABLE awd_attack_logs
    DROP COLUMN IF EXISTS service_id;

ALTER TABLE awd_team_services
    DROP COLUMN IF EXISTS service_id;
