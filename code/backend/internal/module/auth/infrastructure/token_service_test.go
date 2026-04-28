package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	"ctf-platform/pkg/errcode"
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
	if _, err := service.GetSession(context.Background(), session.ID); !errors.Is(err, errcode.ErrUnauthorized) {
		t.Fatalf("expected unauthorized after delete, got %v", err)
	}
}

func TestTokenServiceGetSessionRejectsMissingSession(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	if _, err := service.GetSession(context.Background(), "missing-session"); !errors.Is(err, errcode.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for missing session, got %v", err)
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
