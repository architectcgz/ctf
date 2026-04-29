ALTER TABLE public.contest_awd_services
    DROP CONSTRAINT IF EXISTS contest_awd_services_challenge_id_fkey;

ALTER TABLE public.contest_awd_services
    ALTER COLUMN awd_challenge_id SET NOT NULL;

DROP INDEX IF EXISTS public.uk_contest_awd_services_contest_challenge;

CREATE UNIQUE INDEX IF NOT EXISTS uk_contest_awd_services_contest_awd_challenge
    ON public.contest_awd_services USING btree (contest_id, awd_challenge_id)
    WHERE (deleted_at IS NULL);

ALTER TABLE public.contest_awd_services
    DROP COLUMN IF EXISTS challenge_id;

ALTER TABLE public.awd_team_services
    RENAME COLUMN challenge_id TO awd_challenge_id;

ALTER TABLE public.awd_attack_logs
    RENAME COLUMN challenge_id TO awd_challenge_id;

ALTER TABLE public.awd_traffic_events
    RENAME COLUMN challenge_id TO awd_challenge_id;

ALTER TABLE public.awd_team_services
    DROP CONSTRAINT IF EXISTS awd_team_services_challenge_id_fkey;

ALTER TABLE public.awd_attack_logs
    DROP CONSTRAINT IF EXISTS awd_attack_logs_challenge_id_fkey;

ALTER TABLE public.awd_traffic_events
    DROP CONSTRAINT IF EXISTS awd_traffic_events_challenge_id_fkey;

ALTER TABLE public.awd_team_services
    ADD CONSTRAINT awd_team_services_awd_challenge_id_fkey
    FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;

ALTER TABLE public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_awd_challenge_id_fkey
    FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;

ALTER TABLE public.awd_traffic_events
    ADD CONSTRAINT awd_traffic_events_awd_challenge_id_fkey
    FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;
