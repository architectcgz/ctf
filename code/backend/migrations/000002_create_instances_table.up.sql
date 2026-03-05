CREATE TABLE IF NOT EXISTS instances (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    container_id VARCHAR(64) NOT NULL,
    network_id   VARCHAR(64) DEFAULT NULL,
    status       VARCHAR(16) NOT NULL,
    access_url   VARCHAR(255) DEFAULT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    extend_count INT NOT NULL DEFAULT 0,
    max_extends  INT NOT NULL DEFAULT 2,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_instances_user_id ON instances(user_id);
CREATE INDEX IF NOT EXISTS idx_instances_challenge_id ON instances(challenge_id);
CREATE INDEX IF NOT EXISTS idx_instances_status ON instances(status);
CREATE INDEX IF NOT EXISTS idx_instances_expires_at ON instances(expires_at);
