ALTER TABLE public.images
  ADD COLUMN IF NOT EXISTS digest text,
  ADD COLUMN IF NOT EXISTS source_type text NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS build_job_id bigint,
  ADD COLUMN IF NOT EXISTS last_error text,
  ADD COLUMN IF NOT EXISTS verified_at timestamp without time zone;

CREATE TABLE IF NOT EXISTS public.image_build_jobs (
  id bigserial PRIMARY KEY,
  source_type text NOT NULL,
  challenge_mode text NOT NULL,
  package_slug text NOT NULL,
  source_dir text NOT NULL,
  dockerfile_path text NOT NULL,
  context_path text NOT NULL,
  target_ref text NOT NULL,
  target_digest text,
  status text NOT NULL,
  log_path text,
  error_summary text,
  created_by bigint,
  started_at timestamp without time zone,
  finished_at timestamp without time zone,
  created_at timestamp without time zone,
  updated_at timestamp without time zone
);

CREATE INDEX IF NOT EXISTS idx_image_build_jobs_status_created_at ON public.image_build_jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_image_build_jobs_package_slug ON public.image_build_jobs(package_slug);
CREATE INDEX IF NOT EXISTS idx_images_source_type ON public.images(source_type);
CREATE INDEX IF NOT EXISTS idx_images_build_job_id ON public.images(build_job_id);
