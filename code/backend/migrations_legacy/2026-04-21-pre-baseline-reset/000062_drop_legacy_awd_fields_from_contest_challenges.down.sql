ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_type VARCHAR(32) NOT NULL DEFAULT '';

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_config TEXT NOT NULL DEFAULT '{}';

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_sla_score INTEGER NOT NULL DEFAULT 0;

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_defense_score INTEGER NOT NULL DEFAULT 0;

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_validation_state VARCHAR(24) NOT NULL DEFAULT 'pending';

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_last_preview_at TIMESTAMP NULL;

ALTER TABLE contest_challenges
    ADD COLUMN IF NOT EXISTS awd_checker_last_preview_result TEXT NOT NULL DEFAULT '';

UPDATE contest_challenges AS cc
SET
    awd_checker_type = COALESCE(cas.runtime_config::jsonb ->> 'checker_type', ''),
    awd_checker_config = COALESCE(
        CASE
            WHEN cas.runtime_config::jsonb ? 'checker_config_raw'
                THEN cas.runtime_config::jsonb ->> 'checker_config_raw'
            WHEN cas.runtime_config::jsonb ? 'checker_config'
                THEN (cas.runtime_config::jsonb -> 'checker_config')::text
            ELSE '{}'
        END,
        '{}'
    ),
    awd_sla_score = COALESCE((cas.score_config::jsonb ->> 'awd_sla_score')::INTEGER, 0),
    awd_defense_score = COALESCE((cas.score_config::jsonb ->> 'awd_defense_score')::INTEGER, 0),
    awd_checker_validation_state = COALESCE(cas.awd_checker_validation_state, 'pending'),
    awd_checker_last_preview_at = cas.awd_checker_last_preview_at,
    awd_checker_last_preview_result = COALESCE(cas.awd_checker_last_preview_result, '')
FROM contest_awd_services AS cas
WHERE cas.deleted_at IS NULL
  AND cas.contest_id = cc.contest_id
  AND cas.challenge_id = cc.challenge_id;
