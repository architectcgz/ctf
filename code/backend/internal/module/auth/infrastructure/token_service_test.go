package infrastructure_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
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

	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)
	tokens, err := service.IssueTokens(context.Background(), 42, "alice", "student")
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

	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)
	firstTokens, err := service.IssueTokens(context.Background(), 52, "bob", "student")
	if err != nil {
		t.Fatalf("IssueTokens(first) error = %v", err)
	}
	secondTokens, err := service.IssueTokens(context.Background(), 52, "bob", "student")
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

	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)
	firstTokens, err := service.IssueTokens(context.Background(), 77, "carol", "student")
	if err != nil {
		t.Fatalf("IssueTokens(first) error = %v", err)
	}
	secondTokens, err := service.IssueTokens(context.Background(), 77, "carol", "student")
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

func TestTokenServiceIssueTokensHonorsCancellation(t *testing.T) {
	mini := miniredis.RunT(t)
	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := newTestAuthConfig(t)
	manager, err := jwtpkg.NewManager(cfg, "ctf-platform")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	service := authinfra.NewTokenService(cfg, testWebSocketConfig(), redisClient, manager)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = service.IssueTokens(ctx, 88, "dave", "student")
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

func newTestAuthConfig(t *testing.T) config.AuthConfig {
	t.Helper()

	privateKeyPath, publicKeyPath := writeTestKeyPair(t)
	return config.AuthConfig{
		Issuer:                "ctf-platform-test",
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       24 * time.Hour,
		RefreshCookieName:     "refresh_token",
		RefreshCookiePath:     "/",
		RefreshCookieHTTPOnly: true,
		RefreshCookieSameSite: "lax",
		PrivateKeyPath:        privateKeyPath,
		PublicKeyPath:         publicKeyPath,
		TokenBlacklistPrefix:  "test:blacklist",
	}
}

func writeTestKeyPair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	})

	keyDir := t.TempDir()
	privatePath := filepath.Join(keyDir, "test_private.pem")
	publicPath := filepath.Join(keyDir, "test_public.pem")
	if err := os.WriteFile(privatePath, privatePEM, 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	if err := os.WriteFile(publicPath, publicPEM, 0o644); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return privatePath, publicPath
}
