DROP INDEX IF EXISTS uq_challenges_package_slug;

ALTER TABLE challenges
    DROP COLUMN IF EXISTS package_slug;
