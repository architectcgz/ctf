DROP INDEX IF EXISTS public.idx_contest_realtime_outbox_recipient_user_id;
DROP INDEX IF EXISTS public.idx_contest_realtime_outbox_status_next_attempt;
DROP INDEX IF EXISTS public.uk_contest_realtime_outbox_dedupe_key;
DROP TABLE IF EXISTS public.contest_realtime_outbox;
