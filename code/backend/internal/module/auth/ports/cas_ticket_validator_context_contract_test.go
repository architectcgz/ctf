package ports_test

import (
	"context"

	authports "ctf-platform/internal/module/auth/ports"
)

type ctxOnlyCASTicketValidator struct{}

func (ctxOnlyCASTicketValidator) ValidateTicket(context.Context, string) (*authports.CASPrincipal, error) {
	return nil, nil
}

var _ authports.CASTicketValidator = (*ctxOnlyCASTicketValidator)(nil)
