DROP INDEX IF EXISTS uk_users_teacher_no;
DROP INDEX IF EXISTS uk_users_student_no;

ALTER TABLE users
    DROP COLUMN IF EXISTS teacher_no,
    DROP COLUMN IF EXISTS student_no;
