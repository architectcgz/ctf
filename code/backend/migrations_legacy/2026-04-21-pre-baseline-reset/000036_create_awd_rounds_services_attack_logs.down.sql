DROP INDEX IF EXISTS idx_awd_attack_success;
DROP INDEX IF EXISTS idx_awd_attack_victim;
DROP INDEX IF EXISTS idx_awd_attack_round;
DROP TABLE IF EXISTS awd_attack_logs;

DROP INDEX IF EXISTS idx_awd_ts_team;
DROP INDEX IF EXISTS uk_awd_team_services;
DROP TABLE IF EXISTS awd_team_services;

DROP INDEX IF EXISTS idx_awd_rounds_status;
DROP INDEX IF EXISTS uk_awd_rounds;
DROP TABLE IF EXISTS awd_rounds;
