package commands

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"ctf-platform/pkg/errcode"
)

func mapSubmissionError(err error) error {
	if appErr, ok := err.(*errcode.AppError); ok {
		return appErr
	}
	return errcode.ErrInternal.WithCause(err)
}

func isContestSubmissionUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return pgErr.ConstraintName == "uk_submissions_contest_team_challenge_correct" ||
		pgErr.ConstraintName == "uk_submissions_contest_user_challenge_correct"
}
