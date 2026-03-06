CREATE UNIQUE INDEX IF NOT EXISTS idx_contest_challenges_active_order
    ON contest_challenges(contest_id, "order")
    WHERE deleted_at IS NULL;

ALTER TABLE contest_challenges
    ADD CONSTRAINT fk_contest_challenges_contest
    FOREIGN KEY (contest_id) REFERENCES contests(id) ON DELETE CASCADE;

ALTER TABLE contest_challenges
    ADD CONSTRAINT fk_contest_challenges_challenge
    FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE;
