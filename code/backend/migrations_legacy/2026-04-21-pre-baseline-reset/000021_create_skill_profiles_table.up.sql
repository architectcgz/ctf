CREATE TABLE IF NOT EXISTS skill_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dimension VARCHAR(20) NOT NULL CHECK (
        dimension IN ('web', 'pwn', 'reverse', 'crypto', 'misc', 'forensics')
    ),
    score NUMERIC(5, 4) NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 1),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_skill_profiles_user_dimension
    ON skill_profiles(user_id, dimension);

CREATE INDEX IF NOT EXISTS idx_skill_profiles_user_id
    ON skill_profiles(user_id);

CREATE INDEX IF NOT EXISTS idx_skill_profiles_updated_at
    ON skill_profiles(updated_at);
