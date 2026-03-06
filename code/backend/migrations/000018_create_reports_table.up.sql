CREATE TABLE reports (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(32) NOT NULL,
    format VARCHAR(16) NOT NULL,
    user_id BIGINT DEFAULT NULL,
    class_name VARCHAR(128) DEFAULT NULL,
    status VARCHAR(32) NOT NULL,
    file_path TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ DEFAULT NULL,
    error_msg TEXT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_reports_user_status ON reports(user_id, status, created_at DESC);
CREATE INDEX idx_reports_class_status ON reports(class_name, status, created_at DESC);
CREATE INDEX idx_reports_created_at ON reports(created_at DESC);
