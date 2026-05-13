package ports

import (
	"context"
	"errors"
)

var ErrCASTicketInvalid = errors.New("auth cas ticket invalid")

type CASPrincipal struct {
	Username  string
	Name      string
	Email     string
	ClassName string
	StudentNo string
	TeacherNo string
}

type CASTicketValidator interface {
	ValidateTicket(ctx context.Context, validateURL string) (*CASPrincipal, error)
}
