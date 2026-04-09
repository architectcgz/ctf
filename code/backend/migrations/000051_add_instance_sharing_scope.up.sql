ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS instance_sharing VARCHAR(16) NOT NULL DEFAULT 'per_user';

UPDATE challenges
SET instance_sharing = 'per_user'
WHERE instance_sharing IS NULL
   OR instance_sharing = '';

ALTER TABLE instances
    ADD COLUMN IF NOT EXISTS share_scope VARCHAR(16) NOT NULL DEFAULT 'per_user';

UPDATE instances
SET share_scope = CASE
    WHEN contest_id IS NOT NULL AND team_id IS NOT NULL THEN 'per_team'
    ELSE 'per_user'
END
WHERE share_scope IS NULL
   OR share_scope = '';

DROP INDEX IF EXISTS uk_instances_personal_active;
DROP INDEX IF EXISTS uk_instances_contest_user_active;
DROP INDEX IF EXISTS uk_instances_contest_team_active;

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_personal_active
    ON instances(user_id, challenge_id)
    WHERE contest_id IS NULL
      AND team_id IS NULL
      AND share_scope = 'per_user'
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_contest_user_active
    ON instances(contest_id, user_id, challenge_id)
    WHERE contest_id IS NOT NULL
      AND team_id IS NULL
      AND share_scope = 'per_user'
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_contest_team_active
    ON instances(contest_id, team_id, challenge_id)
    WHERE team_id IS NOT NULL
      AND share_scope = 'per_team'
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_shared_practice_active
    ON instances(challenge_id)
    WHERE contest_id IS NULL
      AND team_id IS NULL
      AND share_scope = 'shared'
      AND status IN ('creating', 'running');

CREATE UNIQUE INDEX IF NOT EXISTS uk_instances_shared_contest_active
    ON instances(contest_id, challenge_id)
    WHERE contest_id IS NOT NULL
      AND team_id IS NULL
      AND share_scope = 'shared'
      AND status IN ('creating', 'running');
