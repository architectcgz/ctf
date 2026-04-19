DROP INDEX IF EXISTS idx_awd_traffic_service;

ALTER TABLE awd_traffic_events
    DROP COLUMN IF EXISTS service_id;
