ALTER TABLE public.contest_awd_services
    ADD COLUMN IF NOT EXISTS challenge_id bigint;

UPDATE public.contest_awd_services
SET challenge_id = awd_challenge_id
WHERE challenge_id IS NULL;

ALTER TABLE public.contest_awd_services
    ALTER COLUMN challenge_id SET NOT NULL;

DROP INDEX IF EXISTS public.uk_contest_awd_services_contest_awd_challenge;

CREATE UNIQUE INDEX IF NOT EXISTS uk_contest_awd_services_contest_challenge
    ON public.contest_awd_services USING btree (contest_id, challenge_id)
    WHERE (deleted_at IS NULL);

ALTER TABLE public.awd_team_services
    DROP CONSTRAINT IF EXISTS awd_team_services_awd_challenge_id_fkey;

ALTER TABLE public.awd_attack_logs
    DROP CONSTRAINT IF EXISTS awd_attack_logs_awd_challenge_id_fkey;

ALTER TABLE public.awd_traffic_events
    DROP CONSTRAINT IF EXISTS awd_traffic_events_awd_challenge_id_fkey;

ALTER TABLE public.awd_team_services
    RENAME COLUMN awd_challenge_id TO challenge_id;

ALTER TABLE public.awd_attack_logs
    RENAME COLUMN awd_challenge_id TO challenge_id;

ALTER TABLE public.awd_traffic_events
    RENAME COLUMN awd_challenge_id TO challenge_id;
