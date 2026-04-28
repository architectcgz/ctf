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
	findByIDWithContextFn            func(ctx context.Context, id int64) (*model.Instance, error)
	findAWDTargetProxyScopeWithCtxFn func(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*runtimeports.AWDTargetProxyScope, error)
	findAWDDefenseSSHScopeWithCtxFn  func(ctx context.Context, userID, contestID, serviceID int64) (*runtimeports.AWDDefenseSSHScope, error)
}

func (s *stubProxyTicketInstanceReader) FindByID(ctx context.Context, id int64) (*model.Instance, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubProxyTicketInstanceReader) FindAWDTargetProxyScope(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*runtimeports.AWDTargetProxyScope, error) {
	if s.findAWDTargetProxyScopeWithCtxFn != nil {
		return s.findAWDTargetProxyScopeWithCtxFn(ctx, userID, contestID, serviceID, victimTeamID)
	}
	return nil, nil
}

func (s *stubProxyTicketInstanceReader) FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*runtimeports.AWDDefenseSSHScope, error) {
	if s.findAWDDefenseSSHScopeWithCtxFn != nil {
		return s.findAWDDefenseSSHScopeWithCtxFn(ctx, userID, contestID, serviceID)
	}
	return nil, nil
}

func TestProxyTicketServiceIssueTicketPersistsClaimsWithTTL(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{}
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Instance, error) {
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

func TestProxyTicketServiceIssueAWDTargetTicketPersistsAttackScope(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{}
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{
		findAWDTargetProxyScopeWithCtxFn: func(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*runtimeports.AWDTargetProxyScope, error) {
			if userID != 1001 || contestID != 3001 || serviceID != 4001 || victimTeamID != 5002 {
				t.Fatalf("unexpected target lookup args: user=%d contest=%d service=%d victim=%d", userID, contestID, serviceID, victimTeamID)
			}
			return &runtimeports.AWDTargetProxyScope{
				InstanceID:     9001,
				ContestID:      contestID,
				AttackerTeamID: 5001,
				VictimTeamID:   victimTeamID,
				ServiceID:      serviceID,
				ChallengeID:    6001,
				ShareScope:     model.InstanceSharingPerTeam,
				AccessURL:      "http://127.0.0.1:39001",
			}, nil
		},
	}, 15*time.Minute)

	ticket, expiresAt, err := service.IssueAWDTargetTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "alice",
		Role:     model.RoleStudent,
	}, 3001, 4001, 5002)
	if err != nil {
		t.Fatalf("IssueAWDTargetTicket() error = %v", err)
	}
	if ticket == "" || expiresAt.IsZero() {
		t.Fatalf("expected issued ticket and expiry, got ticket=%q expires=%s", ticket, expiresAt)
	}
	if store.savedClaims.Purpose != runtimeports.ProxyTicketPurposeAWDAttack {
		t.Fatalf("expected awd attack purpose, got %+v", store.savedClaims)
	}
	if store.savedClaims.InstanceID != 9001 || store.savedClaims.AWDAttackerTeamID == nil || *store.savedClaims.AWDAttackerTeamID != 5001 {
		t.Fatalf("unexpected attacker claims: %+v", store.savedClaims)
	}
	if store.savedClaims.AWDVictimTeamID == nil || *store.savedClaims.AWDVictimTeamID != 5002 {
		t.Fatalf("unexpected victim claims: %+v", store.savedClaims)
	}
	if store.savedClaims.AWDServiceID == nil || *store.savedClaims.AWDServiceID != 4001 {
		t.Fatalf("unexpected service claims: %+v", store.savedClaims)
	}
	if store.savedClaims.AWDChallengeID == nil || *store.savedClaims.AWDChallengeID != 6001 {
		t.Fatalf("unexpected challenge claims: %+v", store.savedClaims)
	}
}

func TestProxyTicketServiceIssueAWDDefenseSSHTicketPersistsOwnTeamScope(t *testing.T) {
	t.Parallel()

	store := &stubProxyTicketStore{}
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{
		findAWDDefenseSSHScopeWithCtxFn: func(ctx context.Context, userID, contestID, serviceID int64) (*runtimeports.AWDDefenseSSHScope, error) {
			if userID != 1001 || contestID != 3001 || serviceID != 4001 {
				t.Fatalf("unexpected defense ssh lookup args: user=%d contest=%d service=%d", userID, contestID, serviceID)
			}
			return &runtimeports.AWDDefenseSSHScope{
				InstanceID:  9001,
				ContestID:   contestID,
				TeamID:      5001,
				ServiceID:   serviceID,
				ChallengeID: 6001,
				ContainerID: "ctr-red-web",
				ShareScope:  model.InstanceSharingPerTeam,
			}, nil
		},
	}, 15*time.Minute)

	ticket, expiresAt, err := service.IssueAWDDefenseSSHTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "alice",
		Role:     model.RoleStudent,
	}, 3001, 4001)
	if err != nil {
		t.Fatalf("IssueAWDDefenseSSHTicket() error = %v", err)
	}
	if ticket == "" || expiresAt.IsZero() {
		t.Fatalf("expected issued ticket and expiry, got ticket=%q expires=%s", ticket, expiresAt)
	}
	if store.savedClaims.Purpose != runtimeports.ProxyTicketPurposeAWDDefenseSSH {
		t.Fatalf("expected awd defense ssh purpose, got %+v", store.savedClaims)
	}
	if store.savedClaims.InstanceID != 9001 {
		t.Fatalf("unexpected instance claims: %+v", store.savedClaims)
	}
	if store.savedClaims.AWDAttackerTeamID == nil || *store.savedClaims.AWDAttackerTeamID != 5001 {
		t.Fatalf("expected own team in attacker team field, got %+v", store.savedClaims)
	}
	if store.savedClaims.AWDVictimTeamID != nil {
		t.Fatalf("defense ssh claims must not include a victim team: %+v", store.savedClaims)
	}
	if store.savedClaims.AWDServiceID == nil || *store.savedClaims.AWDServiceID != 4001 {
		t.Fatalf("unexpected service claims: %+v", store.savedClaims)
	}
	if store.savedClaims.AWDChallengeID == nil || *store.savedClaims.AWDChallengeID != 6001 {
		t.Fatalf("unexpected challenge claims: %+v", store.savedClaims)
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

type proxyTicketContextKey string

func TestProxyTicketServiceIssueTicketPropagatesContextToInstanceReader(t *testing.T) {
	t.Parallel()

	ctxKey := proxyTicketContextKey("proxy-ticket")
	expectedCtxValue := "ctx-proxy-ticket"
	store := &stubProxyTicketStore{}
	readerCalled := false
	service := runtimeqry.NewProxyTicketService(store, &stubProxyTicketInstanceReader{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Instance, error) {
			readerCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected instance reader ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Instance{ID: id, ShareScope: model.InstanceSharingPerUser}, nil
		},
	}, 15*time.Minute)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, _, err := service.IssueTicket(ctx, authctx.CurrentUser{UserID: 1001, Username: "alice", Role: model.RoleStudent}, 2001); err != nil {
		t.Fatalf("IssueTicket() error = %v", err)
	}
	if !readerCalled {
		t.Fatal("expected instance reader to be called")
	}
}
