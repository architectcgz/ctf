package runtime

import (
	"context"
	"fmt"

	"ctf-platform/internal/authctx"
	"ctf-platform/pkg/errcode"
)

type Module struct {
	*Service
	proxyTickets         *ProxyTicketService
	proxyBodyPreviewSize int
}

func NewModule(service *Service, proxyTickets *ProxyTicketService, proxyBodyPreviewSize int) *Module {
	return &Module{
		Service:              service,
		proxyTickets:         proxyTickets,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
	}
}

func (m *Module) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	if m == nil || m.proxyTickets == nil {
		return "", errProxyTicketServiceUnavailable()
	}

	ticket, _, err := m.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (m *Module) ResolveProxyTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error) {
	if m == nil || m.proxyTickets == nil {
		return nil, errProxyTicketServiceUnavailable()
	}
	return m.proxyTickets.ResolveTicket(ctx, ticket)
}

func (m *Module) ProxyTicketMaxAge() int {
	if m == nil || m.proxyTickets == nil || m.proxyTickets.cfg == nil {
		return 0
	}
	return int(m.proxyTickets.cfg.ProxyTicketTTL.Seconds())
}

func (m *Module) ProxyBodyPreviewSize() int {
	if m == nil {
		return 0
	}
	return m.proxyBodyPreviewSize
}

func errProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}
