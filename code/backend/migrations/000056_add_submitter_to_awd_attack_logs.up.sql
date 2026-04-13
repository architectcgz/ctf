ALTER TABLE awd_attack_logs
    ADD COLUMN IF NOT EXISTS submitted_by_user_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_awd_attack_submitter_success
    ON awd_attack_logs(submitted_by_user_id, is_success, score_gained);
