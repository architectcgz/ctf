ALTER TABLE public.instances DROP COLUMN IF EXISTS flag_key_id;

DROP TABLE IF EXISTS public.runtime_cluster_secrets;
