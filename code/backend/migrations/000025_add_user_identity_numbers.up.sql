ALTER TABLE users
    ADD COLUMN IF NOT EXISTS student_no VARCHAR(64) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS teacher_no VARCHAR(64) DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_student_no
    ON users(student_no)
    WHERE deleted_at IS NULL AND student_no IS NOT NULL AND student_no <> '';

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_teacher_no
    ON users(teacher_no)
    WHERE deleted_at IS NULL AND teacher_no IS NOT NULL AND teacher_no <> '';
