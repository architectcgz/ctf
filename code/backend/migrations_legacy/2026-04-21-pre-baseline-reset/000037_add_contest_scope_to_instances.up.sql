ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS contest_id BIGINT,
    ADD COLUMN IF NOT EXISTS team_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_instances_contest_id
    ON instances(contest_id)
    WHERE contest_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_instances_team_id
    ON instances(team_id)
    WHERE team_id IS NOT NULL;
