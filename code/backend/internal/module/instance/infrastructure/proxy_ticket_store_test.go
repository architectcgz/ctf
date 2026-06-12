package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
)

func TestProxyTicketStoreSaveAndFindRoundTrip(t *testing.T) {
	t.Parallel()

	mini := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	store := NewProxyTicketStore(client)
	contestID := int64(3001)
	teamID := int64(5001)
	serviceID := int64(4001)
	challengeID := int64(6001)
	workspaceRevision := int64(7)
	claims := instanceports.ProxyTicketClaims{
		UserID:               1001,
		Username:             "alice",
		Role:                 identitycontracts.RoleStudent,
		InstanceID:           9001,
		ContestID:            &contestID,
		ShareScope:           "per_team",
		Purpose:              instanceports.ProxyTicketPurposeAWDDefenseSSH,
		AWDAttackerTeamID:    &teamID,
		AWDServiceID:         &serviceID,
		AWDChallengeID:       &challengeID,
		AWDWorkspaceRevision: &workspaceRevision,
		IssuedAt:             time.Date(2026, 6, 12, 6, 0, 0, 0, time.UTC),
	}

	if err := store.SaveProxyTicket(context.Background(), "ticket-1", claims, 2*time.Minute); err != nil {
		t.Fatalf("SaveProxyTicket() error = %v", err)
	}

	stored, err := store.FindProxyTicket(context.Background(), "ticket-1")
	if err != nil {
		t.Fatalf("FindProxyTicket() error = %v", err)
	}
	if stored == nil {
		t.Fatal("expected stored claims")
	}
	if stored.UserID != claims.UserID || stored.Username != claims.Username || stored.InstanceID != claims.InstanceID {
		t.Fatalf("unexpected stored claims: %+v", stored)
	}
	if stored.ContestID == nil || *stored.ContestID != contestID {
		t.Fatalf("unexpected contest scope: %+v", stored)
	}
	if stored.AWDWorkspaceRevision == nil || *stored.AWDWorkspaceRevision != workspaceRevision {
		t.Fatalf("unexpected workspace revision: %+v", stored)
	}
	if ttl := mini.TTL(proxyTicketKeyPrefix + ":ticket-1"); ttl <= 0 || ttl > 2*time.Minute {
		t.Fatalf("expected ttl within (0, 2m], got %s", ttl)
	}
}

func TestProxyTicketStoreFindReturnsNilAfterExpiry(t *testing.T) {
	t.Parallel()

	mini := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	store := NewProxyTicketStore(client)
	claims := instanceports.ProxyTicketClaims{
		UserID:     1001,
		Username:   "alice",
		Role:       identitycontracts.RoleStudent,
		InstanceID: 9001,
		ShareScope: "shared",
		Purpose:    instanceports.ProxyTicketPurposeInstanceAccess,
	}

	if err := store.SaveProxyTicket(context.Background(), "ticket-expire", claims, time.Second); err != nil {
		t.Fatalf("SaveProxyTicket() error = %v", err)
	}
	mini.FastForward(2 * time.Second)

	stored, err := store.FindProxyTicket(context.Background(), "ticket-expire")
	if err != nil {
		t.Fatalf("FindProxyTicket() error = %v", err)
	}
	if stored != nil {
		t.Fatalf("expected expired ticket to return nil, got %+v", stored)
	}
}

func TestProxyTicketStoreFindReturnsNilWhenMissing(t *testing.T) {
	t.Parallel()

	mini := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	store := NewProxyTicketStore(client)
	stored, err := store.FindProxyTicket(context.Background(), "missing")
	if err != nil {
		t.Fatalf("FindProxyTicket() error = %v", err)
	}
	if stored != nil {
		t.Fatalf("expected missing ticket to return nil, got %+v", stored)
	}
}

