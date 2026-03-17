package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
	jwtpkg "ctf-platform/pkg/jwt"
)

func TestTokenServiceTracksAndClearsRefreshSession(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig(t)
	manager, err := jwtpkg.NewManager(cfg, "ctf-platform")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	service := NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)
	tokens, err := service.IssueTokens(42, "alice", "student")
	if err != nil {
		t.Fatalf("IssueTokens() error = %v", err)
	}

	refreshClaims, err := service.ParseToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	sessionValue, err := redisClient.Get(context.Background(), rediskeys.TokenKey(42)).Result()
	if err != nil {
		t.Fatalf("load refresh session: %v", err)
	}
	if sessionValue != refreshClaims.ID {
		t.Fatalf("expected refresh session %q, got %q", refreshClaims.ID, sessionValue)
	}

	if err := service.ClearRefreshSession(context.Background(), 42, refreshClaims.ID); err != nil {
		t.Fatalf("ClearRefreshSession() error = %v", err)
	}
	if mini.Exists(rediskeys.TokenKey(42)) {
		t.Fatalf("expected refresh session key to be removed")
	}
}

func TestTokenServiceClearRefreshSessionDoesNotRemoveNewerSession(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig(t)
	manager, err := jwtpkg.NewManager(cfg, "ctf-platform")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	service := NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)
	firstTokens, err := service.IssueTokens(52, "bob", "student")
	if err != nil {
		t.Fatalf("IssueTokens(first) error = %v", err)
	}
	secondTokens, err := service.IssueTokens(52, "bob", "student")
	if err != nil {
		t.Fatalf("IssueTokens(second) error = %v", err)
	}

	firstClaims, err := service.ParseToken(firstTokens.RefreshToken)
	if err != nil {
		t.Fatalf("ParseToken(first) error = %v", err)
	}
	secondClaims, err := service.ParseToken(secondTokens.RefreshToken)
	if err != nil {
		t.Fatalf("ParseToken(second) error = %v", err)
	}

	if err := service.ClearRefreshSession(context.Background(), 52, firstClaims.ID); err != nil {
		t.Fatalf("ClearRefreshSession() error = %v", err)
	}

	sessionValue, err := redisClient.Get(context.Background(), rediskeys.TokenKey(52)).Result()
	if err != nil {
		t.Fatalf("load refresh session: %v", err)
	}
	if sessionValue != secondClaims.ID {
		t.Fatalf("expected newer refresh session %q, got %q", secondClaims.ID, sessionValue)
	}
}

func TestTokenServiceRefreshAccessTokenRejectsStaleRefreshSession(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig(t)
	manager, err := jwtpkg.NewManager(cfg, "ctf-platform")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	service := NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)
	firstTokens, err := service.IssueTokens(77, "carol", "student")
	if err != nil {
		t.Fatalf("IssueTokens(first) error = %v", err)
	}
	secondTokens, err := service.IssueTokens(77, "carol", "student")
	if err != nil {
		t.Fatalf("IssueTokens(second) error = %v", err)
	}

	if _, err := service.RefreshAccessToken(context.Background(), firstTokens.RefreshToken); !errors.Is(err, errcode.ErrTokenRevoked) {
		t.Fatalf("expected stale refresh token to be revoked, got %v", err)
	}

	payload, err := service.RefreshAccessToken(context.Background(), secondTokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshAccessToken(second) error = %v", err)
	}
	if payload == nil || payload.AccessToken == "" {
		t.Fatalf("expected refreshed access token, got %+v", payload)
	}
}

func TestTokenServiceIssueTokensWithContextHonorsCancellation(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig(t)
	manager, err := jwtpkg.NewManager(cfg, "ctf-platform")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	service := NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = service.IssueTokensWithContext(ctx, 88, "dave", "student")
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func testWebSocketConfig() config.WebSocketConfig {
	return config.WebSocketConfig{
		TicketTTL:       time.Minute,
		TicketKeyPrefix: "test:ws:ticket",
	}
}
