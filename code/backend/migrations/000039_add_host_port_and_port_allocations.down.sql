DROP INDEX IF EXISTS uk_instances_active_host_port;
DROP INDEX IF EXISTS uk_instances_contest_team_active;
DROP INDEX IF EXISTS uk_instances_contest_user_active;
DROP INDEX IF EXISTS uk_instances_personal_active;
DROP INDEX IF EXISTS idx_port_allocations_instance_id;
DROP TABLE IF EXISTS port_allocations;

ALTER TABLE instances
    DROP COLUMN IF EXISTS host_port;
