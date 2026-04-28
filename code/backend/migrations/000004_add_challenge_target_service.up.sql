ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS target_protocol varchar(16) DEFAULT 'http' NOT NULL;

ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS target_port integer DEFAULT 0 NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_challenges_target_protocol'
    ) THEN
        ALTER TABLE challenges
            ADD CONSTRAINT chk_challenges_target_protocol
            CHECK (target_protocol IN ('http', 'tcp'));
    END IF;
END $$;
