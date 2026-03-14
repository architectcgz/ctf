DROP INDEX IF EXISTS uk_submissions_contest_team_challenge_correct;
DROP INDEX IF EXISTS uk_submissions_contest_user_challenge_correct;

CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_user_challenge_correct
    ON submissions(contest_id, user_id, challenge_id)
    WHERE is_correct = TRUE
      AND contest_id IS NOT NULL;
