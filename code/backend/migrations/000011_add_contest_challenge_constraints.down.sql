-- 删除外键约束
ALTER TABLE contest_challenges DROP FOREIGN KEY fk_contest_challenges_challenge;
ALTER TABLE contest_challenges DROP FOREIGN KEY fk_contest_challenges_contest;

-- 删除 order 唯一索引
ALTER TABLE contest_challenges DROP INDEX idx_contest_order;
