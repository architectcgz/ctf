UPDATE users
SET email = NULL
WHERE email IS NOT NULL
  AND btrim(email) = '';

DROP INDEX IF EXISTS uk_users_email;

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_email
  ON users(email)
  WHERE deleted_at IS NULL
    AND email IS NOT NULL
    AND btrim(email) <> '';
