CREATE TABLE IF NOT EXISTS public.runtime_cluster_secrets (
    name character varying(128) NOT NULL,
    active_key_id character varying(128) NOT NULL,
    active_fingerprint character varying(128) NOT NULL,
    key_fingerprints jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT runtime_cluster_secrets_pkey PRIMARY KEY (name)
);

ALTER TABLE public.instances ADD COLUMN IF NOT EXISTS flag_key_id character varying(128);
