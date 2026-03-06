-- 添加 nonce 字段到 instances 表
ALTER TABLE instances
ADD COLUMN nonce VARCHAR(64);
