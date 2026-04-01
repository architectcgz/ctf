CREATE TABLE IF NOT EXISTS awd_traffic_events (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    round_id BIGINT NOT NULL REFERENCES awd_rounds(id) ON DELETE CASCADE,
    attacker_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    victim_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    challenge_id BIGINT NOT NULL REFERENCES challenges(id) ON DELETE RESTRICT,
    method VARCHAR(16) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    status_code INT NOT NULL,
    source VARCHAR(32) NOT NULL DEFAULT 'runtime_proxy',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_awd_traffic_round_created ON awd_traffic_events(round_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_awd_traffic_round_summary ON awd_traffic_events(round_id, method, path, status_code);
CREATE INDEX IF NOT EXISTS idx_awd_traffic_attacker ON awd_traffic_events(round_id, attacker_team_id);
CREATE INDEX IF NOT EXISTS idx_awd_traffic_victim ON awd_traffic_events(round_id, victim_team_id);
