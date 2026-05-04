DROP INDEX IF EXISTS public.uk_instances_active_host_port;

CREATE UNIQUE INDEX uk_instances_active_host_port
ON public.instances USING btree (host_port)
WHERE host_port > 0
  AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text]));
