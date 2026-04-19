package application_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
	runtimeqry "ctf-platform/internal/module/runtime/application/queries"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type stubProxyTicketStore struct {
	savedTicket string
	savedClaims runtimeports.ProxyTicketClaims
	savedTTL    time.Duration
	findClaims  *runtimeports.ProxyTicketClaims
	saveErr     error
	findErr     error
}

func (s *stubProxyTicketStore) SaveProxyTicket(ctx context.Context, ticket string, claims runtimeports.ProxyTicketClaims, ttl time.Duration) error {
	s.savedTicket = ticket
	s.savedClaims = claims
	s.savedTTL = ttl
	return s.saveErr
}

func (s *stubProxyTicketStore) FindProxyTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	return s.findClaims, s.findErr
}

type stubProxyTicketInstanceReader struct {
	findByIDFn func(id int64) (*model.Instance, error)
}

func (s *stubProxyTicketInstanceReader) FindByID(id int64) (*model.Instance, error) {
	if s.findByIDFn == nil {
		return nil, nil
	}
	return s.findByIDFn(id)
}

func TestProxyTicketServiceIssueTicketPersistsClaimsWithTTL(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{}
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{
		findByIDFn: func(id int64) (*model.Instance, error) {
			contestID := int64(3001)
			return &model.Instance{
				ID:          id,
				ContestID:   &contestID,
				ChallengeID: 2001,
				ShareScope:  model.InstanceSharingShared,
			}, nil
		},
	}, 15*time.Minute)

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
	if store.savedClaims.ContestID == nil || *store.savedClaims.ContestID != 3001 {
		t.Fatalf("unexpected saved contest scope: %+v", store.savedClaims)
	}
	if store.savedClaims.ShareScope != model.InstanceSharingShared {
		t.Fatalf("unexpected saved share scope: %+v", store.savedClaims)
	}
	if store.savedTTL != 15*time.Minute {
		t.Fatalf("saved ttl = %s, want %s", store.savedTTL, 15*time.Minute)
	}
	if service.MaxAge() != 900 {
		t.Fatalf("MaxAge() = %d, want 900", service.MaxAge())
	}
}

func TestProxyTicketServiceResolveTicketAllowsClaimsWithoutChallengeID(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{
		findClaims: &runtimeports.ProxyTicketClaims{
			UserID:     1001,
			Username:   "alice",
			Role:       model.RoleStudent,
			InstanceID: 2001,
			ShareScope: model.InstanceSharingPerTeam,
		},
	}
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{}, 15*time.Minute)

	claims, err := service.ResolveTicket(context.Background(), "ticket-1")
	if err != nil {
		t.Fatalf("ResolveTicket() error = %v", err)
	}
	if claims == nil || claims.InstanceID != 2001 || claims.UserID != 1001 {
		t.Fatalf("unexpected resolved claims: %+v", claims)
	}
}

func TestProxyTicketServiceResolveTicketRejectsInvalidClaims(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{
		findClaims: &runtimeports.ProxyTicketClaims{
			UserID:     1001,
			InstanceID: 2001,
		},
	}
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{}, 15*time.Minute)

	_, err := service.ResolveTicket(context.Background(), "ticket-1")
	if err == nil || err.Error() != errcode.ErrProxyTicketInvalid.Error() {
		t.Fatalf("expected invalid ticket error, got %v", err)
	}
}
