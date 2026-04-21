ALTER TABLE contest_challenges
    DROP COLUMN IF EXISTS first_blood_by,
    DROP COLUMN IF EXISTS contest_score;

ALTER TABLE teams
    DROP COLUMN IF EXISTS last_solve_at,
    DROP COLUMN IF EXISTS total_score;
