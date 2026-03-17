ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS host_port INT;

CREATE TABLE IF NOT EXISTS port_allocations (
    port        INT PRIMARY KEY,
    instance_id BIGINT NULL REFERENCES instances(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_port_allocations_instance_id
    ON port_allocations(instance_id);

UPDATE instances
SET host_port = CAST(substring(access_url from ':(\d+)(?:/)?$') AS INTEGER)
WHERE host_port IS NULL
  AND access_url IS NOT NULL
  AND access_url <> ''
  AND substring(access_url from ':(\d+)(?:/)?$') IS NOT NULL;

INSERT INTO port_allocations (port, instance_id, created_at, updated_at)
SELECT host_port, id, NOW(), NOW()
FROM instances
WHERE host_port IS NOT NULL
  AND status IN ('creating', 'running')
ON CONFLICT (port) DO NOTHING;

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_personal_active
    ON instances(user_id, challenge_id)
    WHERE contest_id IS NULL
      AND team_id IS NULL
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_contest_user_active
    ON instances(contest_id, user_id, challenge_id)
    WHERE contest_id IS NOT NULL
      AND team_id IS NULL
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_contest_team_active
    ON instances(contest_id, team_id, challenge_id)
    WHERE team_id IS NOT NULL
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_active_host_port
    ON instances(host_port)
    WHERE host_port IS NOT NULL
      AND status IN ('creating', 'running');
