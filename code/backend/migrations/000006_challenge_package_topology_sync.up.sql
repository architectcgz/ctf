CREATE TABLE public.challenge_package_revisions (
    id bigint NOT NULL,
    challenge_id bigint NOT NULL,
    revision_no integer NOT NULL,
    source_type character varying(32) NOT NULL,
    parent_revision_id bigint,
    package_slug character varying(128) NOT NULL,
    archive_path text NOT NULL DEFAULT ''::text,
    source_dir text NOT NULL DEFAULT ''::text,
    manifest_snapshot text NOT NULL DEFAULT ''::text,
    topology_source_path character varying(255) NOT NULL DEFAULT ''::character varying,
    topology_snapshot text NOT NULL DEFAULT ''::text,
    created_by bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE SEQUENCE public.challenge_package_revisions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.challenge_package_revisions_id_seq OWNED BY public.challenge_package_revisions.id;
ALTER TABLE ONLY public.challenge_package_revisions ALTER COLUMN id SET DEFAULT nextval('public.challenge_package_revisions_id_seq'::regclass);

ALTER TABLE public.challenge_topologies
    ADD COLUMN source_type character varying(32) NOT NULL DEFAULT 'platform_manual'::character varying,
    ADD COLUMN source_path character varying(255) NOT NULL DEFAULT ''::character varying,
    ADD COLUMN package_revision_id bigint,
    ADD COLUMN package_baseline_spec jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN sync_status character varying(32) NOT NULL DEFAULT 'clean'::character varying,
    ADD COLUMN last_export_revision_id bigint;

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_challenge_package_revisions_challenge_revision
    ON public.challenge_package_revisions USING btree (challenge_id, revision_no);

CREATE INDEX idx_challenge_topologies_package_revision_id
    ON public.challenge_topologies USING btree (package_revision_id);

CREATE INDEX idx_challenge_topologies_last_export_revision_id
    ON public.challenge_topologies USING btree (last_export_revision_id);

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_parent_revision_id_fkey FOREIGN KEY (parent_revision_id) REFERENCES public.challenge_package_revisions(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.challenge_package_revisions
    ADD CONSTRAINT challenge_package_revisions_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_package_revision_id_fkey FOREIGN KEY (package_revision_id) REFERENCES public.challenge_package_revisions(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.challenge_topologies
    ADD CONSTRAINT challenge_topologies_last_export_revision_id_fkey FOREIGN KEY (last_export_revision_id) REFERENCES public.challenge_package_revisions(id) ON DELETE SET NULL;
