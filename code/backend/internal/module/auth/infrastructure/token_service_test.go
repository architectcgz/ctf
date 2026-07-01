package infrastructure_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
)

type failDelHook struct{}

func (failDelHook) DialHook(next redislib.DialHook) redislib.DialHook {
	return next
}

func (failDelHook) ProcessHook(next redislib.ProcessHook) redislib.ProcessHook {
	return func(ctx context.Context, cmd redislib.Cmder) error {
		if cmd.Name() == "del" {
			return errors.New("forced del failure")
		}
		return next(ctx, cmd)
	}
}

func (failDelHook) ProcessPipelineHook(next redislib.ProcessPipelineHook) redislib.ProcessPipelineHook {
	return next
}

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

func TestTokenServiceRevokeAllUserSessionsInvalidatesLegacySessionWithoutIndex(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig()
	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient)

	payload := `{"id":"legacy-session","user_id":42,"username":"alice","role":"student","expires_at":"2099-01-01T00:00:00Z"}`
	if err := redisClient.Set(context.Background(), cfg.SessionKeyPrefix+":legacy-session", payload, time.Hour).Err(); err != nil {
		t.Fatalf("seed legacy payload: %v", err)
	}

	if err := service.RevokeAllUserSessions(context.Background(), 42); err != nil {
		t.Fatalf("RevokeAllUserSessions() error = %v", err)
	}

	if _, err := service.GetSession(context.Background(), "legacy-session"); !errors.Is(err, authcontracts.ErrAccessTokenExpired) {
		t.Fatalf("expected legacy session to be revoked after user-wide revocation, got %v", err)
	}
}

func TestTokenServiceRevokeAllUserSessionsIgnoresCleanupDeleteFailureAfterVersionIncrement(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	redisClient.AddHook(failDelHook{})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)
	session, err := service.CreateSession(context.Background(), 42, "alice", "student")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if err := service.RevokeAllUserSessions(context.Background(), 42); err != nil {
		t.Fatalf("expected cleanup delete failure to be ignored, got %v", err)
	}
	if _, err := service.GetSession(context.Background(), session.ID); !errors.Is(err, authcontracts.ErrAccessTokenExpired) {
		t.Fatalf("expected session to be invalidated after version increment, got %v", err)
	}
}

func TestTokenServiceListUserSessionsReturnsActiveSessions(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	s1, err := service.CreateSession(context.Background(), 42, "alice", "student")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	s2, err := service.CreateSession(context.Background(), 42, "alice", "student")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	sessions, err := service.ListUserSessions(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListUserSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(sessions))
	}

	ids := map[string]bool{}
	for _, s := range sessions {
		ids[s.ID] = true
	}
	if !ids[s1.ID] || !ids[s2.ID] {
		t.Fatalf("expected both session ids in result, got %+v", sessions)
	}
}

func TestTokenServiceListUserSessionsEmptyForUnknownUser(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	sessions, err := service.ListUserSessions(context.Background(), 999)
	if err != nil {
		t.Fatalf("ListUserSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions for unknown user, got %d", len(sessions))
	}
}

func TestTokenServiceListUserSessionsFiltersExpired(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig()
	cfg.SessionTTL = 1 * time.Hour
	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient)

	_, err := service.CreateSession(context.Background(), 42, "alice", "student")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Fast-forward redis time to expire the session
	mini.FastForward(2 * time.Hour)

	sessions, err := service.ListUserSessions(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListUserSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions after expiry, got %d", len(sessions))
	}
}

func TestTokenServiceListUserSessionsRejectsNilContext(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := authinfra.NewTokenService(newTestAuthConfig(), testWebSocketConfig(), redisClient)

	if _, err := service.ListUserSessions(nil, 42); err == nil {
		t.Fatal("expected ListUserSessions() to reject nil context")
	}
}

func TestTokenServiceLegacyMCPTokenSurfaceRemoved(t *testing.T) {
	files := []string{
		"token_service.go",
		"../contracts/token_service.go",
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		// 旧 MCP token 入口一旦残留，就会形成绕过 OAuth browser authorization 的第二条认证路径。
		for _, legacy := range []string{"MCPToken", "IssueMCPToken", "ResolveMCPToken", "mcpTokenPayload", "mcpTokenKey"} {
			if strings.Contains(string(content), legacy) {
				t.Fatalf("%s still contains legacy MCP token surface %q", file, legacy)
			}
		}
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
		OAuth: config.AuthOAuthConfig{
			AccessTokenTTL: time.Hour,
		},
	}
}
