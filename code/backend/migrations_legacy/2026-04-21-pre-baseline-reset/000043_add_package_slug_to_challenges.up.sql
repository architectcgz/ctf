ALTER TABLE challenges
    ADD COLUMN IF NOT EXISTS package_slug VARCHAR(128) DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_challenges_package_slug
    ON challenges(package_slug);
