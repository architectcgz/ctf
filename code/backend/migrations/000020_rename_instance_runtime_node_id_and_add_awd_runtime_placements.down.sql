DROP INDEX IF EXISTS public.idx_contest_runtime_placements_runtime_node_status;
DROP INDEX IF EXISTS public.idx_contest_runtime_placements_active_contest;
DROP TABLE IF EXISTS public.contest_runtime_placements;

DROP INDEX IF EXISTS public.idx_instances_runtime_node_container_id;
DROP INDEX IF EXISTS public.idx_instances_runtime_node_network_id;

ALTER TABLE public.instances DROP CONSTRAINT IF EXISTS instances_runtime_node_id_fkey;

ALTER TABLE public.instances RENAME COLUMN runtime_node_id TO node_id;

ALTER INDEX IF EXISTS public.idx_instances_runtime_node_id RENAME TO idx_instances_node_id;

ALTER TABLE public.instances
    ADD CONSTRAINT instances_node_id_fkey
    FOREIGN KEY (node_id) REFERENCES public.runtime_nodes(id);
