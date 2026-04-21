DROP INDEX IF EXISTS idx_awd_attack_submitter_success;

ALTER TABLE awd_attack_logs
    DROP COLUMN IF EXISTS submitted_by_user_id;
