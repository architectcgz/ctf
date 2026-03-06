CREATE TABLE IF NOT EXISTS skill_profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    dimension VARCHAR(20) NOT NULL CHECK (dimension != ''),
    score DOUBLE NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 1),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_user_dimension (user_id, dimension),
    KEY idx_user_id (user_id),
    KEY idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
