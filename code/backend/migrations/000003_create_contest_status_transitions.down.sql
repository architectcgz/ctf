DROP INDEX IF EXISTS public.idx_contest_status_transitions_occurred_at;
DROP INDEX IF EXISTS public.uk_contest_status_transitions_contest_version;
DROP TABLE IF EXISTS public.contest_status_transitions;
ALTER TABLE public.contests DROP COLUMN IF EXISTS status_version;
