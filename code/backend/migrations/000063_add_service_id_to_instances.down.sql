DROP INDEX IF EXISTS idx_instances_service_id;

ALTER TABLE instances
    DROP COLUMN IF EXISTS service_id;
