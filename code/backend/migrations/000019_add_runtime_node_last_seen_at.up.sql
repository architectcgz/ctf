ALTER TABLE public.runtime_nodes
    ADD COLUMN last_seen_at timestamp with time zone;

CREATE INDEX idx_runtime_nodes_health_schedulable_seen
    ON public.runtime_nodes USING btree (schedulable, health_status, last_seen_at);
