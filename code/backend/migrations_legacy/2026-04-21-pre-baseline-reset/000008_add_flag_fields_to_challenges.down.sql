-- 回滚 Flag 字段
DROP INDEX IF EXISTS idx_challenges_flag_type;

ALTER TABLE challenges
DROP COLUMN flag_type,
DROP COLUMN flag_hash,
DROP COLUMN flag_salt;
