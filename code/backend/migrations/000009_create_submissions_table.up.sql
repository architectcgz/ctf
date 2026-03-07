-- 创建 submissions 表
CREATE TABLE IF NOT EXISTS submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    flag VARCHAR(500),
    is_correct BOOLEAN NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_submissions_user_challenge ON submissions(user_id, challenge_id);
CREATE INDEX idx_submissions_submitted_at ON submissions(submitted_at);
