CREATE TABLE IF NOT EXISTS skill_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    dimension VARCHAR(50) NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_skill_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_dimension ON skill_profiles(user_id, dimension);
CREATE INDEX idx_skill_profiles_updated_at ON skill_profiles(updated_at);
