ALTER TABLE awd_attack_logs
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'legacy';

CREATE INDEX IF NOT EXISTS idx_awd_attack_source ON awd_attack_logs(round_id, source);
