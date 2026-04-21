DROP INDEX IF EXISTS idx_instances_team_id;
DROP INDEX IF EXISTS idx_instances_contest_id;

ALTER TABLE instances
    DROP COLUMN IF EXISTS team_id,
    DROP COLUMN IF EXISTS contest_id;
