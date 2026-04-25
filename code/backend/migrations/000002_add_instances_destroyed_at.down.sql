DROP INDEX IF EXISTS public.idx_instances_destroyed_at;

ALTER TABLE public.instances
    DROP COLUMN IF EXISTS destroyed_at;
