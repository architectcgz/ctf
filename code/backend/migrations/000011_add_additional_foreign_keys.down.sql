ALTER TABLE ONLY public.submissions
    DROP CONSTRAINT IF EXISTS submissions_team_id_fkey;

ALTER TABLE ONLY public.submissions
    DROP CONSTRAINT IF EXISTS submissions_contest_id_fkey;

ALTER TABLE ONLY public.submissions
    DROP CONSTRAINT IF EXISTS submissions_challenge_id_fkey;

ALTER TABLE ONLY public.submissions
    DROP CONSTRAINT IF EXISTS submissions_user_id_fkey;

ALTER TABLE ONLY public.instances
    DROP CONSTRAINT IF EXISTS instances_service_id_fkey;

ALTER TABLE ONLY public.instances
    DROP CONSTRAINT IF EXISTS instances_team_id_fkey;

ALTER TABLE ONLY public.instances
    DROP CONSTRAINT IF EXISTS instances_contest_id_fkey;

ALTER TABLE ONLY public.instances
    DROP CONSTRAINT IF EXISTS instances_challenge_id_fkey;

ALTER TABLE ONLY public.instances
    DROP CONSTRAINT IF EXISTS instances_user_id_fkey;

ALTER TABLE ONLY public.reports
    DROP CONSTRAINT IF EXISTS reports_user_id_fkey;

ALTER TABLE ONLY public.awd_service_operations
    DROP CONSTRAINT IF EXISTS awd_service_operations_requested_by_id_fkey;

ALTER TABLE ONLY public.team_members
    DROP CONSTRAINT IF EXISTS team_members_contest_id_fkey;

ALTER TABLE ONLY public.awd_service_templates
    DROP CONSTRAINT IF EXISTS awd_service_templates_created_by_fkey;

ALTER TABLE ONLY public.awd_service_templates
    DROP CONSTRAINT IF EXISTS awd_service_templates_last_verified_by_fkey;

ALTER TABLE ONLY public.awd_challenges
    DROP CONSTRAINT IF EXISTS awd_challenges_created_by_fkey;

ALTER TABLE ONLY public.awd_challenges
    DROP CONSTRAINT IF EXISTS awd_challenges_last_verified_by_fkey;

ALTER TABLE ONLY public.awd_attack_logs
    DROP CONSTRAINT IF EXISTS awd_attack_logs_submitted_by_user_id_fkey;

ALTER TABLE ONLY public.audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey;
