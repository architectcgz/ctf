DROP INDEX IF EXISTS public.idx_runtime_nodes_health_schedulable_seen;

ALTER TABLE public.runtime_nodes
    DROP COLUMN IF EXISTS last_seen_at;
