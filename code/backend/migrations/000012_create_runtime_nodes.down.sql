DROP INDEX IF EXISTS idx_instances_node_id;

ALTER TABLE instances
    DROP COLUMN IF EXISTS node_id;

DROP TABLE IF EXISTS runtime_nodes;
