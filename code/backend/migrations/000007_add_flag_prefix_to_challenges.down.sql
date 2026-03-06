-- 回滚 flag_prefix 字段
ALTER TABLE challenges
DROP COLUMN IF EXISTS flag_prefix;
