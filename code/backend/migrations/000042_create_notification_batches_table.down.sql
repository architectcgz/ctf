DROP INDEX IF EXISTS uk_notifications_batch_user;
DROP INDEX IF EXISTS idx_notifications_batch_id;
DROP INDEX IF EXISTS idx_notifications_user_created_at;
DROP INDEX IF EXISTS idx_notification_batches_created_by_created_at;

ALTER TABLE notifications
    DROP COLUMN IF EXISTS batch_id;

DROP TABLE IF EXISTS notification_batches;
