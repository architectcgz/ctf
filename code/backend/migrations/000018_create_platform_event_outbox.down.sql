DROP INDEX IF EXISTS idx_platform_event_outbox_locked_until;
DROP INDEX IF EXISTS idx_platform_event_outbox_pending_due;
DROP INDEX IF EXISTS idx_platform_event_outbox_dedupe_key;
DROP TABLE IF EXISTS platform_event_outbox;
DROP INDEX IF EXISTS idx_notifications_source_event_key;
ALTER TABLE notifications DROP COLUMN IF EXISTS source_event_key;
