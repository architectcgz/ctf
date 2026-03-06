CREATE TABLE IF NOT EXISTS contest_challenges (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    contest_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    points INT NOT NULL DEFAULT 0,
    `order` INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY idx_contest_challenge (contest_id, challenge_id, deleted_at),
    KEY idx_contest_id (contest_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
