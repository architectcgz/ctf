package queries

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

// ProxyTicketService 保留 runtime compat import path，并把访问票据能力委托给 instance owner。
type ProxyTicketService struct {
	service   instancecontracts.ProxyTicketService
	ticketTTL int
}

func NewProxyTicketService(service instancecontracts.ProxyTicketService, ticketTTL int) *ProxyTicketService {
	return &ProxyTicketService{
		service:   service,
		ticketTTL: ticketTTL,
	}
}

func (s *ProxyTicketService) IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error) {
	if s == nil || s.service == nil {
		return "", time.Time{}, errRuntimeCompatProxyTicketServiceUnavailable()
	}
	return s.service.IssueTicket(ctx, user, instanceID)
}

func (s *ProxyTicketService) IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error) {
	if s == nil || s.service == nil {
		return "", time.Time{}, errRuntimeCompatProxyTicketServiceUnavailable()
	}
	return s.service.IssueAWDTargetTicket(ctx, user, contestID, serviceID, victimTeamID)
}

func (s *ProxyTicketService) IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (string, time.Time, error) {
	if s == nil || s.service == nil {
		return "", time.Time{}, errRuntimeCompatProxyTicketServiceUnavailable()
	}
	return s.service.IssueAWDDefenseSSHTicket(ctx, user, contestID, serviceID)
}

func (s *ProxyTicketService) ResolveAWDTargetAccessURL(ctx context.Context, claims *runtimeports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
	if s == nil || s.service == nil {
		return "", errRuntimeCompatProxyTicketServiceUnavailable()
	}
	return s.service.ResolveAWDTargetAccessURL(ctx, claims, contestID, serviceID, victimTeamID)
}

func (s *ProxyTicketService) ResolveTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if s == nil || s.service == nil {
		return nil, errRuntimeCompatProxyTicketServiceUnavailable()
	}
	return s.service.ResolveTicket(ctx, ticket)
}

func (s *ProxyTicketService) MaxAge() int {
	if s == nil {
		return 0
	}
	return s.ticketTTL
}

func errRuntimeCompatProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("runtime proxy ticket compat service is not configured"))
}
