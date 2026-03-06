-- B27: 竞赛提交与计分功能

-- 扩展 submissions 表
ALTER TABLE submissions
ADD COLUMN contest_id BIGINT DEFAULT NULL,
ADD COLUMN team_id BIGINT DEFAULT NULL,
ADD COLUMN score INT NOT NULL DEFAULT 0;

CREATE INDEX idx_submissions_contest_id ON submissions(contest_id) WHERE contest_id IS NOT NULL;

-- 创建 contests 表
CREATE TABLE contests (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    team_mode BOOLEAN NOT NULL DEFAULT FALSE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contests_status ON contests(status);

-- 创建 teams 表
CREATE TABLE teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    contest_id BIGINT NOT NULL,
    captain_id BIGINT NOT NULL,
    total_score INT NOT NULL DEFAULT 0,
    last_solve_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teams_contest_id ON teams(contest_id);
CREATE INDEX idx_teams_contest_score ON teams(contest_id, total_score DESC, last_solve_at ASC);

-- 创建 contest_registrations 表
CREATE TABLE contest_registrations (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    team_id BIGINT DEFAULT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uk_contest_reg_user ON contest_registrations(contest_id, user_id);

-- 创建 contest_challenges 表
CREATE TABLE contest_challenges (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    challenge_id BIGINT NOT NULL,
    contest_score INT DEFAULT NULL,
    first_blood_by BIGINT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contest_challenges_contest ON contest_challenges(contest_id);
CREATE UNIQUE INDEX uk_contest_challenge ON contest_challenges(contest_id, challenge_id);
