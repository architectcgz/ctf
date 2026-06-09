CREATE INDEX IF NOT EXISTS idx_instances_status_updated_id
    ON public.instances USING btree (status, updated_at, id);
