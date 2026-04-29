ALTER TABLE ONLY public.challenge_topologies DROP CONSTRAINT IF EXISTS challenge_topologies_last_export_revision_id_fkey;
ALTER TABLE ONLY public.challenge_topologies DROP CONSTRAINT IF EXISTS challenge_topologies_package_revision_id_fkey;
ALTER TABLE ONLY public.challenge_package_revisions DROP CONSTRAINT IF EXISTS challenge_package_revisions_created_by_fkey;
ALTER TABLE ONLY public.challenge_package_revisions DROP CONSTRAINT IF EXISTS challenge_package_revisions_parent_revision_id_fkey;
ALTER TABLE ONLY public.challenge_package_revisions DROP CONSTRAINT IF EXISTS challenge_package_revisions_challenge_id_fkey;

DROP INDEX IF EXISTS public.idx_challenge_topologies_last_export_revision_id;
DROP INDEX IF EXISTS public.idx_challenge_topologies_package_revision_id;
DROP INDEX IF EXISTS public.idx_challenge_package_revisions_challenge_revision;

ALTER TABLE public.challenge_topologies
    DROP COLUMN IF EXISTS last_export_revision_id,
    DROP COLUMN IF EXISTS sync_status,
    DROP COLUMN IF EXISTS package_baseline_spec,
    DROP COLUMN IF EXISTS package_revision_id,
    DROP COLUMN IF EXISTS source_path,
    DROP COLUMN IF EXISTS source_type;

DROP TABLE IF EXISTS public.challenge_package_revisions;
DROP SEQUENCE IF EXISTS public.challenge_package_revisions_id_seq;
