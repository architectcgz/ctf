ALTER TABLE awd_team_services
    ADD COLUMN IF NOT EXISTS service_id BIGINT;

ALTER TABLE awd_attack_logs
    ADD COLUMN IF NOT EXISTS service_id BIGINT;

UPDATE awd_team_services ats
SET service_id = cas.id
FROM awd_rounds ar
JOIN contest_awd_services cas
  ON cas.contest_id = ar.contest_id
 AND cas.challenge_id = ats.challenge_id
 AND cas.deleted_at IS NULL
WHERE ats.round_id = ar.id
  AND ats.service_id IS NULL;

UPDATE awd_attack_logs aal
SET service_id = cas.id
FROM awd_rounds ar
JOIN contest_awd_services cas
  ON cas.contest_id = ar.contest_id
 AND cas.challenge_id = aal.challenge_id
 AND cas.deleted_at IS NULL
WHERE aal.round_id = ar.id
  AND aal.service_id IS NULL;

DELETE FROM awd_team_services
WHERE service_id IS NULL;

DELETE FROM awd_attack_logs
WHERE service_id IS NULL;

ALTER TABLE awd_team_services
    ALTER COLUMN service_id SET NOT NULL;

ALTER TABLE awd_attack_logs
    ALTER COLUMN service_id SET NOT NULL;

DROP INDEX IF EXISTS uk_awd_team_services;
CREATE UNIQUE INDEX IF NOT EXISTS uk_awd_team_services
    ON awd_team_services(round_id, team_id, service_id);

CREATE INDEX IF NOT EXISTS idx_awd_ts_round_team_service
    ON awd_team_services(round_id, team_id, service_id);

CREATE INDEX IF NOT EXISTS idx_awd_attack_round_service_success
    ON awd_attack_logs(round_id, attacker_team_id, victim_team_id, service_id, is_success);
