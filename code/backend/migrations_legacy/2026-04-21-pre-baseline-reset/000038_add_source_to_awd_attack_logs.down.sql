DROP INDEX IF EXISTS idx_awd_attack_source;

ALTER TABLE awd_attack_logs
    DROP COLUMN IF EXISTS source;
