CREATE TABLE IF NOT EXISTS user_scores (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_score INT NOT NULL DEFAULT 0,
    solved_count INT NOT NULL DEFAULT 0,
    rank INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_scores_rank ON user_scores(rank);
CREATE INDEX idx_user_scores_total_score ON user_scores(total_score DESC);
