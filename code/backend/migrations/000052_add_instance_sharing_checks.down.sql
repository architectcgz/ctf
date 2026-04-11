ALTER TABLE instances
    DROP CONSTRAINT IF EXISTS chk_instances_share_scope;

ALTER TABLE challenges
    DROP CONSTRAINT IF EXISTS chk_challenges_instance_sharing;
