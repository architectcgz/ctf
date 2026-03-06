CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'vulnerability',
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_tags_name_type ON tags(name, type);

CREATE TABLE IF NOT EXISTS challenge_tags (
    id BIGSERIAL PRIMARY KEY,
    challenge_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_challenge_tags ON challenge_tags(challenge_id, tag_id);
CREATE INDEX idx_challenge_tags_tag_id ON challenge_tags(tag_id);
