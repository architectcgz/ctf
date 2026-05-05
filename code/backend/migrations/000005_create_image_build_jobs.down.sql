DROP INDEX IF EXISTS public.idx_images_build_job_id;
DROP INDEX IF EXISTS public.idx_images_source_type;
DROP INDEX IF EXISTS public.idx_image_build_jobs_package_slug;
DROP INDEX IF EXISTS public.idx_image_build_jobs_status_created_at;

DROP TABLE IF EXISTS public.image_build_jobs;

ALTER TABLE public.images
  DROP COLUMN IF EXISTS verified_at,
  DROP COLUMN IF EXISTS last_error,
  DROP COLUMN IF EXISTS build_job_id,
  DROP COLUMN IF EXISTS source_type,
  DROP COLUMN IF EXISTS digest;
