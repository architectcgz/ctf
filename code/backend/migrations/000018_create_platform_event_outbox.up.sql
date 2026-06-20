CREATE TABLE IF NOT EXISTS public.platform_event_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_name TEXT NOT NULL,
    payload BYTEA NOT NULL,
    payload_version INTEGER NOT NULL,
    route TEXT NOT NULL,
    dedupe_key TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    locked_by TEXT NOT NULL DEFAULT '',
    locked_until TIMESTAMPTZ NULL,
    stream_message_id TEXT NOT NULL DEFAULT '',
    dispatched_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_event_outbox_dedupe_key
    ON public.platform_event_outbox (dedupe_key)
    WHERE dedupe_key <> '';

CREATE INDEX IF NOT EXISTS idx_platform_event_outbox_pending_due
    ON public.platform_event_outbox (status, next_attempt_at, id)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_platform_event_outbox_locked_until
    ON public.platform_event_outbox (locked_until)
    WHERE status = 'pending';

ALTER TABLE public.notifications
    ADD COLUMN IF NOT EXISTS source_event_key TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_source_event_key
    ON public.notifications (source_event_key)
    WHERE source_event_key <> '';
