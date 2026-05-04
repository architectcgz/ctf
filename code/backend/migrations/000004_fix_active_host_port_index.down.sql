DROP INDEX IF EXISTS public.uk_instances_active_host_port;

CREATE UNIQUE INDEX uk_instances_active_host_port
ON public.instances USING btree (host_port)
WHERE host_port IS NOT NULL
  AND ((status)::text = ANY (ARRAY[('creating'::character varying)::text, ('running'::character varying)::text]));
