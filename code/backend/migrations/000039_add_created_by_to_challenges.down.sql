DROP INDEX IF EXISTS idx_challenges_created_by;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS created_by;
