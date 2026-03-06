-- 添加部分唯一索引：防止同一用户对同一题目重复提交正确答案
CREATE UNIQUE INDEX idx_submissions_user_challenge_correct
ON submissions(user_id, challenge_id)
WHERE is_correct = true;
