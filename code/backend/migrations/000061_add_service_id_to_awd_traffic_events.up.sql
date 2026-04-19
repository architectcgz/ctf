ALTER TABLE awd_traffic_events
    ADD COLUMN IF NOT EXISTS service_id BIGINT;

UPDATE awd_traffic_events ate
SET service_id = cas.id
FROM contest_awd_services cas
WHERE cas.contest_id = ate.contest_id
  AND cas.challenge_id = ate.challenge_id
  AND ate.service_id IS NULL;

DELETE FROM awd_traffic_events
WHERE service_id IS NULL;

ALTER TABLE awd_traffic_events
    ALTER COLUMN service_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_awd_traffic_service
    ON awd_traffic_events(service_id);
