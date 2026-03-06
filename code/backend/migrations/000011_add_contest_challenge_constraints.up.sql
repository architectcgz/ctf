-- 添加 order 唯一索引
ALTER TABLE contest_challenges ADD UNIQUE KEY idx_contest_order (contest_id, `order`);

-- 添加外键约束
ALTER TABLE contest_challenges
ADD CONSTRAINT fk_contest_challenges_contest
FOREIGN KEY (contest_id) REFERENCES contests(id) ON DELETE CASCADE;

ALTER TABLE contest_challenges
ADD CONSTRAINT fk_contest_challenges_challenge
FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE;
