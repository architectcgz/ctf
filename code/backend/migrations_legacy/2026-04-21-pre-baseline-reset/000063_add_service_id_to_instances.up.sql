ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS service_id BIGINT;

UPDATE instances AS i
SET service_id = cas.id
FROM contests AS co,
     contest_awd_services AS cas
WHERE i.contest_id = co.id
  AND cas.contest_id = co.id
  AND cas.challenge_id = i.challenge_id
  AND cas.deleted_at IS NULL
  AND co.mode = 'awd'
  AND i.service_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_instances_service_id
    ON instances(service_id)
    WHERE service_id IS NOT NULL;
