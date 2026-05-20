package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
)

func TestTokenServiceCreateGetAndDeleteSession(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	session, err := service.CreateSession(context.Background(), 42, "alice", "student")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID == "" {
		t.Fatal("expected non-empty session id")
	}

	loaded, err := service.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if loaded.UserID != 42 || loaded.Username != "alice" || loaded.Role != "student" {
		t.Fatalf("unexpected session payload: %+v", loaded)
	}

	if err := service.DeleteSession(context.Background(), session.ID); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := service.GetSession(context.Background(), session.ID); !errors.Is(err, authcontracts.ErrAccessTokenExpired) {
		t.Fatalf("expected access token expired after delete, got %v", err)
	}
}

func TestTokenServiceGetSessionRejectsMissingSession(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	if _, err := service.GetSession(context.Background(), "missing-session"); !errors.Is(err, authcontracts.ErrAccessTokenExpired) {
		t.Fatalf("expected access token expired for missing session, got %v", err)
	}
}

func TestTokenServiceGetSessionRejectsEmptySessionID(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	if _, err := service.GetSession(context.Background(), ""); !errors.Is(err, authcontracts.ErrTokenInvalid) {
		t.Fatalf("expected token invalid for empty session id, got %v", err)
	}
}

func TestTokenServiceGetSessionRejectsCorruptedPayload(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig()
	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient)

	if err := redisClient.Set(context.Background(), cfg.SessionKeyPrefix+":broken-session", "{bad-json", time.Hour).Err(); err != nil {
		t.Fatalf("seed corrupted payload: %v", err)
	}

	if _, err := service.GetSession(context.Background(), "broken-session"); !errors.Is(err, authcontracts.ErrTokenInvalid) {
		t.Fatalf("expected token invalid for corrupted payload, got %v", err)
	}
}

func TestTokenServiceGetSessionRejectsExpiredPayload(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig()
	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient)

	payload := `{"id":"expired-session","user_id":42,"username":"alice","role":"student","expires_at":"2000-01-01T00:00:00Z"}`
	if err := redisClient.Set(context.Background(), cfg.SessionKeyPrefix+":expired-session", payload, time.Hour).Err(); err != nil {
		t.Fatalf("seed expired payload: %v", err)
	}

	if _, err := service.GetSession(context.Background(), "expired-session"); !errors.Is(err, authcontracts.ErrAccessTokenExpired) {
		t.Fatalf("expected access token expired for expired payload, got %v", err)
	}
}

func TestTokenServiceCreateSessionRejectsNilContext(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	if _, err := service.CreateSession(nil, 42, "alice", "student"); err == nil {
		t.Fatal("expected CreateSession() to reject nil context")
	}
}

func testWebSocketConfig() config.WebSocketConfig {
	return config.WebSocketConfig{
		TicketTTL:       time.Minute,
		TicketKeyPrefix: "test:ws:ticket",
	}
}

func newTestAuthConfig() config.AuthConfig {
	return config.AuthConfig{
		SessionTTL:            24 * time.Hour,
		SessionCookieName:     "ctf_session",
		SessionCookiePath:     "/",
		SessionCookieHTTPOnly: true,
		SessionCookieSameSite: "lax",
		SessionKeyPrefix:      "test:session",
	}
}
