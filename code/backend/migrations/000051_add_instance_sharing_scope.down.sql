DROP INDEX IF EXISTS uk_instances_shared_contest_active;
DROP INDEX IF EXISTS uk_instances_shared_practice_active;
DROP INDEX IF EXISTS uk_instances_contest_team_active;
DROP INDEX IF EXISTS uk_instances_contest_user_active;
DROP INDEX IF EXISTS uk_instances_personal_active;

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

ALTER TABLE instances
    DROP COLUMN IF EXISTS share_scope;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS instance_sharing;
