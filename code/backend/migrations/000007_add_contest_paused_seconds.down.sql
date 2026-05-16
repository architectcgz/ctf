ALTER TABLE public.contests
DROP COLUMN IF EXISTS runtime_recovery_applied_seconds,
DROP COLUMN IF EXISTS runtime_recovery_key,
DROP COLUMN IF EXISTS paused_seconds;
