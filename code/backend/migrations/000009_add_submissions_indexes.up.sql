-- 优化完成人数查询
CREATE INDEX IF NOT EXISTS idx_submissions_challenge_correct ON submissions(challenge_id, is_correct);

-- 优化用户完成状态查询（覆盖索引）
CREATE INDEX IF NOT EXISTS idx_submissions_user_challenge_correct ON submissions(user_id, challenge_id, is_correct);
