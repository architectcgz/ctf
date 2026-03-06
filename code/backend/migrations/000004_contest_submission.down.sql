-- 回滚 B27: 竞赛提交与计分功能

DROP TABLE IF EXISTS contest_challenges;
DROP TABLE IF EXISTS contest_registrations;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS contests;

DROP INDEX IF EXISTS idx_submissions_contest_id;

ALTER TABLE submissions
DROP COLUMN IF EXISTS contest_id,
DROP COLUMN IF EXISTS team_id,
DROP COLUMN IF EXISTS score;
