CREATE TABLE IF NOT EXISTS awd_rounds (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    round_number INT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ DEFAULT NULL,
    ended_at TIMESTAMPTZ DEFAULT NULL,
    attack_score INT NOT NULL DEFAULT 50,
    defense_score INT NOT NULL DEFAULT 50,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_awd_rounds ON awd_rounds(contest_id, round_number);
CREATE INDEX IF NOT EXISTS idx_awd_rounds_status ON awd_rounds(contest_id, status);

CREATE TABLE IF NOT EXISTS awd_team_services (
    id BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES awd_rounds(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    challenge_id BIGINT NOT NULL REFERENCES challenges(id) ON DELETE RESTRICT,
    service_status VARCHAR(16) NOT NULL DEFAULT 'up',
    check_result JSONB NOT NULL DEFAULT '{}'::jsonb,
    attack_received INT NOT NULL DEFAULT 0,
    defense_score INT NOT NULL DEFAULT 0,
    attack_score INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_awd_team_services ON awd_team_services(round_id, team_id, challenge_id);
CREATE INDEX IF NOT EXISTS idx_awd_ts_team ON awd_team_services(team_id, round_id);

CREATE TABLE IF NOT EXISTS awd_attack_logs (
    id BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES awd_rounds(id) ON DELETE CASCADE,
    attacker_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    victim_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    challenge_id BIGINT NOT NULL REFERENCES challenges(id) ON DELETE RESTRICT,
    attack_type VARCHAR(32) NOT NULL,
    submitted_flag VARCHAR(512) DEFAULT NULL,
    is_success BOOLEAN NOT NULL DEFAULT FALSE,
    score_gained INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_awd_attack_round ON awd_attack_logs(round_id, attacker_team_id);
CREATE INDEX IF NOT EXISTS idx_awd_attack_victim ON awd_attack_logs(round_id, victim_team_id);
CREATE INDEX IF NOT EXISTS idx_awd_attack_success ON awd_attack_logs(round_id, is_success) WHERE is_success = TRUE;
