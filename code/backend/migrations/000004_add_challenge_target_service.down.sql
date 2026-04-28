ALTER TABLE challenges
    DROP CONSTRAINT IF EXISTS chk_challenges_target_protocol;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS target_port;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS target_protocol;
