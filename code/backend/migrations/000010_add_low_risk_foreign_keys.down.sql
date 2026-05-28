ALTER TABLE ONLY public.teams
    DROP CONSTRAINT IF EXISTS teams_captain_id_fkey;

ALTER TABLE ONLY public.team_members
    DROP CONSTRAINT IF EXISTS team_members_user_id_fkey;

ALTER TABLE ONLY public.images
    DROP CONSTRAINT IF EXISTS images_build_job_id_fkey;

ALTER TABLE ONLY public.contest_registrations
    DROP CONSTRAINT IF EXISTS contest_registrations_reviewed_by_fkey;

ALTER TABLE ONLY public.contest_registrations
    DROP CONSTRAINT IF EXISTS contest_registrations_team_id_fkey;

ALTER TABLE ONLY public.contest_registrations
    DROP CONSTRAINT IF EXISTS contest_registrations_user_id_fkey;

ALTER TABLE ONLY public.contest_registrations
    DROP CONSTRAINT IF EXISTS contest_registrations_contest_id_fkey;

ALTER TABLE ONLY public.contest_awd_services
    DROP CONSTRAINT IF EXISTS contest_awd_services_awd_challenge_id_fkey;

ALTER TABLE ONLY public.contest_awd_services
    DROP CONSTRAINT IF EXISTS contest_awd_services_contest_id_fkey;

ALTER TABLE ONLY public.challenge_tags
    DROP CONSTRAINT IF EXISTS challenge_tags_tag_id_fkey;

ALTER TABLE ONLY public.challenge_tags
    DROP CONSTRAINT IF EXISTS challenge_tags_challenge_id_fkey;

ALTER TABLE ONLY public.challenge_hints
    DROP CONSTRAINT IF EXISTS challenge_hints_challenge_id_fkey;
