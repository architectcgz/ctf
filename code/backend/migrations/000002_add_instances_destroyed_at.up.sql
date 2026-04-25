ALTER TABLE public.instances
    ADD COLUMN destroyed_at timestamp with time zone;

UPDATE public.instances
SET destroyed_at = updated_at
WHERE destroyed_at IS NULL
  AND status IN ('stopped', 'expired');

CREATE INDEX idx_instances_destroyed_at ON public.instances USING btree (destroyed_at);
