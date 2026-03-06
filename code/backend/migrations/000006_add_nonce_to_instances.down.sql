-- 回滚 nonce 字段
ALTER TABLE instances
DROP COLUMN IF EXISTS nonce;
