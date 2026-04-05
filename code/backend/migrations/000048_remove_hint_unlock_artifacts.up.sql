ALTER TABLE challenge_hints
    DROP COLUMN IF EXISTS cost_points;

DROP INDEX IF EXISTS idx_challenge_hint_unlocks_user_challenge;
DROP INDEX IF EXISTS uk_challenge_hint_unlocks_user_hint;
DROP TABLE IF EXISTS challenge_hint_unlocks;
