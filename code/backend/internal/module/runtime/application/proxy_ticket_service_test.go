package application

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/pkg/errcode"
)

type stubProxyTicketStore struct {
	savedTicket string
	savedClaims ProxyTicketClaims
	savedTTL    time.Duration
	findClaims  *ProxyTicketClaims
	saveErr     error
	findErr     error
}

func (s *stubProxyTicketStore) SaveProxyTicket(ctx context.Context, ticket string, claims ProxyTicketClaims, ttl time.Duration) error {
	s.savedTicket = ticket
	s.savedClaims = claims
	s.savedTTL = ttl
	return s.saveErr
}

func (s *stubProxyTicketStore) FindProxyTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error) {
	return s.findClaims, s.findErr
}

func TestProxyTicketServiceIssueTicketPersistsClaimsWithTTL(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{}
	service := NewProxyTicketService(store, 15*time.Minute)

	ticket, expiresAt, err := service.IssueTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "alice",
		Role:     "student",
	}, 2001)
	if err != nil {
		t.Fatalf("IssueTicket() error = %v", err)
	}
	if ticket == "" {
		t.Fatal("expected issued ticket")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected expires at")
	}
	if store.savedTicket != ticket {
		t.Fatalf("expected stored ticket %q, got %q", ticket, store.savedTicket)
	}
	if store.savedClaims.UserID != 1001 || store.savedClaims.InstanceID != 2001 {
		t.Fatalf("unexpected saved claims: %+v", store.savedClaims)
	}
	if store.savedTTL != 15*time.Minute {
		t.Fatalf("saved ttl = %s, want %s", store.savedTTL, 15*time.Minute)
	}
	if service.MaxAge() != 900 {
		t.Fatalf("MaxAge() = %d, want 900", service.MaxAge())
	}
}

func TestProxyTicketServiceResolveTicketRejectsInvalidClaims(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{
		findClaims: &ProxyTicketClaims{
			UserID:     1001,
			InstanceID: 2001,
		},
	}
	service := NewProxyTicketService(store, 15*time.Minute)

	_, err := service.ResolveTicket(context.Background(), "ticket-1")
	if err == nil || err.Error() != errcode.ErrProxyTicketInvalid.Error() {
		t.Fatalf("expected invalid ticket error, got %v", err)
	}
}
