ALTER TABLE challenges
    DROP CONSTRAINT IF EXISTS chk_challenges_instance_sharing;

ALTER TABLE challenges
    ADD CONSTRAINT chk_challenges_instance_sharing
    CHECK (instance_sharing IN ('per_user', 'per_team', 'shared'));

ALTER TABLE instances
    DROP CONSTRAINT IF EXISTS chk_instances_share_scope;

ALTER TABLE instances
    ADD CONSTRAINT chk_instances_share_scope
    CHECK (share_scope IN ('per_user', 'per_team', 'shared'));
