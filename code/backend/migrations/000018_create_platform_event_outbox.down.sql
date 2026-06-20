DROP INDEX IF EXISTS public.idx_platform_event_outbox_locked_until;
DROP INDEX IF EXISTS public.idx_platform_event_outbox_pending_due;
DROP INDEX IF EXISTS public.idx_platform_event_outbox_dedupe_key;
DROP TABLE IF EXISTS public.platform_event_outbox;
DROP INDEX IF EXISTS public.idx_notifications_source_event_key;
ALTER TABLE public.notifications DROP COLUMN IF EXISTS source_event_key;
