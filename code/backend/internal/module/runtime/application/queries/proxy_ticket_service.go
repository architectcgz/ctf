package queries

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type ProxyTicketService struct {
	store     runtimeports.ProxyTicketStore
	ticketTTL time.Duration
}

func NewProxyTicketService(store runtimeports.ProxyTicketStore, ticketTTL time.Duration) *ProxyTicketService {
	return &ProxyTicketService{
		store:     store,
		ticketTTL: ticketTTL,
	}
}

func (s *ProxyTicketService) IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error) {
	if s == nil || s.store == nil || s.ticketTTL <= 0 {
		return "", time.Time{}, errProxyTicketServiceUnavailable()
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	claims := runtimeports.ProxyTicketClaims{
		UserID:     user.UserID,
		Username:   user.Username,
		Role:       user.Role,
		InstanceID: instanceID,
		IssuedAt:   time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) ResolveTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if s == nil || s.store == nil {
		return nil, errProxyTicketServiceUnavailable()
	}
	if ticket == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	claims, err := s.store.FindProxyTicket(ctx, ticket)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if claims == nil {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if claims.UserID <= 0 || claims.InstanceID <= 0 || claims.Username == "" || claims.Role == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	return claims, nil
}

func (s *ProxyTicketService) MaxAge() int {
	if s == nil || s.ticketTTL <= 0 {
		return 0
	}
	return int(s.ticketTTL.Seconds())
}

func errProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

func generateProxyToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
