WITH missing_team_user_ids AS (
    SELECT DISTINCT t.captain_id AS user_id
    FROM public.teams t
    LEFT JOIN public.users u ON u.id = t.captain_id
    WHERE t.captain_id IS NOT NULL AND u.id IS NULL
    UNION
    SELECT DISTINCT tm.user_id AS user_id
    FROM public.team_members tm
    LEFT JOIN public.users u ON u.id = tm.user_id
    WHERE tm.user_id IS NOT NULL AND u.id IS NULL
)
INSERT INTO public.users (
    id,
    username,
    password_hash,
    email,
    role,
    status,
    name,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    m.user_id,
    'ghost_user_' || m.user_id,
    '!historical-placeholder!',
    'ghost+' || m.user_id || '@invalid.local',
    'student',
    'inactive',
    '历史占位用户-' || m.user_id,
    now(),
    now(),
    now()
FROM missing_team_user_ids m;

ALTER TABLE ONLY public.challenge_hints
    ADD CONSTRAINT challenge_hints_challenge_id_fkey
    FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.challenge_tags
    ADD CONSTRAINT challenge_tags_challenge_id_fkey
    FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.challenge_tags
    ADD CONSTRAINT challenge_tags_tag_id_fkey
    FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.contest_awd_services
    ADD CONSTRAINT contest_awd_services_contest_id_fkey
    FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.contest_awd_services
    ADD CONSTRAINT contest_awd_services_awd_challenge_id_fkey
    FOREIGN KEY (awd_challenge_id) REFERENCES public.awd_challenges(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.contest_registrations
    ADD CONSTRAINT contest_registrations_contest_id_fkey
    FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.contest_registrations
    ADD CONSTRAINT contest_registrations_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.contest_registrations
    ADD CONSTRAINT contest_registrations_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.contest_registrations
    ADD CONSTRAINT contest_registrations_reviewed_by_fkey
    FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_build_job_id_fkey
    FOREIGN KEY (build_job_id) REFERENCES public.image_build_jobs(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_captain_id_fkey
    FOREIGN KEY (captain_id) REFERENCES public.users(id) ON DELETE RESTRICT;
