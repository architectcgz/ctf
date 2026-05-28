ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.awd_attack_logs
    ADD CONSTRAINT awd_attack_logs_submitted_by_user_id_fkey
    FOREIGN KEY (submitted_by_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.awd_challenges
    ADD CONSTRAINT awd_challenges_last_verified_by_fkey
    FOREIGN KEY (last_verified_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.awd_challenges
    ADD CONSTRAINT awd_challenges_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.awd_service_templates
    ADD CONSTRAINT awd_service_templates_last_verified_by_fkey
    FOREIGN KEY (last_verified_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.awd_service_templates
    ADD CONSTRAINT awd_service_templates_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_contest_id_fkey
    FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.awd_service_operations
    ADD CONSTRAINT awd_service_operations_requested_by_id_fkey
    FOREIGN KEY (requested_by_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.instances
    ADD CONSTRAINT instances_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.instances
    ADD CONSTRAINT instances_challenge_id_fkey
    FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.instances
    ADD CONSTRAINT instances_contest_id_fkey
    FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.instances
    ADD CONSTRAINT instances_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.instances
    ADD CONSTRAINT instances_service_id_fkey
    FOREIGN KEY (service_id) REFERENCES public.contest_awd_services(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_challenge_id_fkey
    FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_contest_id_fkey
    FOREIGN KEY (contest_id) REFERENCES public.contests(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES public.teams(id) ON DELETE SET NULL;
