ALTER TABLE public.contests
ADD COLUMN paused_seconds bigint DEFAULT 0 NOT NULL,
ADD COLUMN runtime_recovery_key varchar(191) DEFAULT '' NOT NULL,
ADD COLUMN runtime_recovery_applied_seconds bigint DEFAULT 0 NOT NULL;
