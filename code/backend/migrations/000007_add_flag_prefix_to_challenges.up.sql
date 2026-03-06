-- 添加 flag_prefix 字段到 challenges 表
ALTER TABLE challenges
ADD COLUMN flag_prefix VARCHAR(32) DEFAULT 'flag';
