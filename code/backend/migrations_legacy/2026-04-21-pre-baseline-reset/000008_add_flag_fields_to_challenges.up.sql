-- 添加 Flag 相关字段到 challenges 表
ALTER TABLE challenges
ADD COLUMN flag_type VARCHAR(16) NOT NULL DEFAULT 'static',
ADD COLUMN flag_hash VARCHAR(128),
ADD COLUMN flag_salt VARCHAR(64);

-- 添加索引用于查询优化
CREATE INDEX idx_challenges_flag_type ON challenges(flag_type);
