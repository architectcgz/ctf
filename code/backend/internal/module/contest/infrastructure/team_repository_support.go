package infrastructure

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *TeamRepository) IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, constraint)
	}
	return false
}
