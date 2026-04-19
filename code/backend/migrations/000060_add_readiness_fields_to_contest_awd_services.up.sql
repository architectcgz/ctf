ALTER TABLE contest_awd_services
    ADD COLUMN IF NOT EXISTS awd_checker_validation_state VARCHAR(24) NOT NULL DEFAULT 'pending';

ALTER TABLE contest_awd_services
    ADD COLUMN IF NOT EXISTS awd_checker_last_preview_at TIMESTAMP NULL;

ALTER TABLE contest_awd_services
    ADD COLUMN IF NOT EXISTS awd_checker_last_preview_result TEXT NOT NULL DEFAULT '';

UPDATE contest_awd_services AS cas
SET
    awd_checker_validation_state = cc.awd_checker_validation_state,
    awd_checker_last_preview_at = cc.awd_checker_last_preview_at,
    awd_checker_last_preview_result = cc.awd_checker_last_preview_result
FROM contest_challenges AS cc
WHERE cc.contest_id = cas.contest_id
  AND cc.challenge_id = cas.challenge_id
  AND cas.deleted_at IS NULL;
