ALTER TABLE contest_challenges DROP CONSTRAINT IF EXISTS fk_contest_challenges_challenge;
ALTER TABLE contest_challenges DROP CONSTRAINT IF EXISTS fk_contest_challenges_contest;
DROP INDEX IF EXISTS idx_contest_challenges_active_order;
