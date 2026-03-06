-- 创建队伍表
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    captain_id BIGINT NOT NULL,
    invite_code VARCHAR(6) NOT NULL,
    max_members INT NOT NULL DEFAULT 4,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_teams_contest_id ON teams(contest_id);
CREATE INDEX IF NOT EXISTS idx_teams_captain_id ON teams(captain_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_teams_invite_code ON teams(invite_code);
CREATE UNIQUE INDEX IF NOT EXISTS uk_teams_contest_name ON teams(contest_id, name) WHERE deleted_at IS NULL;
ALTER TABLE teams
    ADD CONSTRAINT fk_teams_contest
    FOREIGN KEY (contest_id) REFERENCES contests(id) ON DELETE CASCADE;

-- 创建队伍成员表
CREATE TABLE IF NOT EXISTS team_members (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_team_members_team_user UNIQUE(team_id, user_id),
    CONSTRAINT uk_team_members_contest_user UNIQUE(contest_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
ALTER TABLE team_members
    ADD CONSTRAINT fk_team_members_team
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
