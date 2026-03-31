CREATE TABLE notification_batches (
    id              BIGSERIAL PRIMARY KEY,
    created_by      BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    type            VARCHAR(16) NOT NULL,
    title           VARCHAR(256) NOT NULL,
    content         TEXT NOT NULL,
    link            VARCHAR(512) DEFAULT NULL,
    audience_mode   VARCHAR(16) NOT NULL,
    audience_rules  JSONB NOT NULL DEFAULT '{}'::jsonb,
    recipient_count INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE notifications
    ADD COLUMN batch_id BIGINT NULL REFERENCES notification_batches(id) ON DELETE SET NULL;

CREATE INDEX idx_notification_batches_created_by_created_at
    ON notification_batches(created_by, created_at DESC);
CREATE INDEX idx_notifications_user_created_at
    ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_batch_id
    ON notifications(batch_id);
CREATE UNIQUE INDEX uk_notifications_batch_user
    ON notifications(batch_id, user_id)
    WHERE batch_id IS NOT NULL;
