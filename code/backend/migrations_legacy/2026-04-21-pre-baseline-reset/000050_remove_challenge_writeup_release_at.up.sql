DROP INDEX IF EXISTS idx_challenge_writeups_visibility_release;

ALTER TABLE challenge_writeups
    DROP COLUMN IF EXISTS release_at;
