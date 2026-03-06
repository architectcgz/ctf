CREATE TABLE IF NOT EXISTS contest_registrations (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    team_id BIGINT,
    status VARCHAR(16) NOT NULL DEFAULT 'approved',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_contest_reg_user
    ON contest_registrations(contest_id, user_id);
