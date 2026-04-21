DROP INDEX IF EXISTS uk_submissions_contest_user_challenge_correct;

WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY contest_id, team_id, challenge_id
               ORDER BY submitted_at ASC, id ASC
           ) AS rn
    FROM submissions
    WHERE is_correct = TRUE
      AND contest_id IS NOT NULL
      AND team_id IS NOT NULL
)
DELETE FROM submissions s
USING ranked r
WHERE s.id = r.id
  AND r.rn > 1;

UPDATE teams t
SET total_score = COALESCE(scores.total_score, 0),
    last_solve_at = scores.last_solve_at
FROM (
    SELECT team_id,
           SUM(score) AS total_score,
           MAX(submitted_at) AS last_solve_at
    FROM submissions
    WHERE is_correct = TRUE
      AND contest_id IS NOT NULL
      AND team_id IS NOT NULL
    GROUP BY team_id
) scores
WHERE t.id = scores.team_id;

UPDATE teams
SET total_score = 0,
    last_solve_at = NULL
WHERE id NOT IN (
    SELECT DISTINCT team_id
    FROM submissions
    WHERE is_correct = TRUE
      AND contest_id IS NOT NULL
      AND team_id IS NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_user_challenge_correct
    ON submissions(contest_id, user_id, challenge_id)
    WHERE is_correct = TRUE
      AND contest_id IS NOT NULL
      AND team_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_team_challenge_correct
    ON submissions(contest_id, team_id, challenge_id)
    WHERE is_correct = TRUE
      AND contest_id IS NOT NULL
      AND team_id IS NOT NULL;
