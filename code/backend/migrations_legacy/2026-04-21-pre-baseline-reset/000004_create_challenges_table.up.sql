CREATE TABLE IF NOT EXISTS challenges (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(128) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(32) NOT NULL,
    difficulty VARCHAR(16) NOT NULL,
    points INT NOT NULL DEFAULT 0,
    image_id BIGINT DEFAULT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_challenges_status ON challenges(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_challenges_category ON challenges(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_challenges_difficulty ON challenges(difficulty) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_challenges_image_id ON challenges(image_id);
