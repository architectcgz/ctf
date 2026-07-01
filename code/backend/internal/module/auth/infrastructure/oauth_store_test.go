package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

func TestOAuthStoreClientAndConsentLifecycle(t *testing.T) {
	ctx := context.Background()
	store, _ := newOAuthStoreTestEnv(t)
	client := authcontracts.OAuthClient{
		ClientID:                "client_123",
		ClientName:              "local-codex",
		ClientURI:               "https://codex.example.test",
		RedirectURIs:            []string{"http://127.0.0.1:14567/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   authcontracts.OAuthScopeMCPChallengeRead,
		TokenEndpointAuthMethod: "none",
	}

	if err := store.SaveClient(ctx, client); err != nil {
		t.Fatalf("SaveClient() error = %v", err)
	}
	loaded, err := store.FindClientByID(ctx, "client_123")
	if err != nil {
		t.Fatalf("FindClientByID() error = %v", err)
	}
	if loaded == nil || loaded.ClientName != "local-codex" || len(loaded.RedirectURIs) != 1 {
		t.Fatalf("unexpected loaded client: %+v", loaded)
	}
	if !loaded.AllowsRedirectURI("http://127.0.0.1:14567/callback") {
		t.Fatalf("registered client should allow exact redirect uri: %+v", loaded.RedirectURIs)
	}
	if loaded.AllowsRedirectURI("http://127.0.0.1:14567/callback/extra") {
		t.Fatalf("registered client must require exact redirect uri match")
	}

	consent := authcontracts.OAuthConsent{
		UserID:   42,
		ClientID: "client_123",
		Scope:    authcontracts.OAuthScopeMCPChallengeRead,
	}
	if err := store.UpsertConsent(ctx, consent); err != nil {
		t.Fatalf("UpsertConsent() error = %v", err)
	}
	active, err := store.FindActiveConsent(ctx, 42, "client_123", authcontracts.OAuthScopeMCPChallengeRead)
	if err != nil {
		t.Fatalf("FindActiveConsent() error = %v", err)
	}
	if active == nil || active.UserID != 42 {
		t.Fatalf("expected active consent, got %+v", active)
	}
	if err := store.RevokeConsent(ctx, 42, "client_123", authcontracts.OAuthScopeMCPChallengeRead); err != nil {
		t.Fatalf("RevokeConsent() error = %v", err)
	}
	active, err = store.FindActiveConsent(ctx, 42, "client_123", authcontracts.OAuthScopeMCPChallengeRead)
	if err != nil {
		t.Fatalf("FindActiveConsent() after revoke error = %v", err)
	}
	if active != nil {
		t.Fatalf("expected revoked consent to be hidden, got %+v", active)
	}
}

func TestOAuthStoreAuthorizationCodeIsSingleUseAndHashed(t *testing.T) {
	ctx := context.Background()
	store, mini := newOAuthStoreTestEnv(t)
	claims := authcontracts.OAuthAuthorizationCode{
		UserID:              42,
		Username:            "alice",
		Role:                "student",
		ClientID:            "client_123",
		RedirectURI:         "http://127.0.0.1:14567/callback",
		Scope:               authcontracts.OAuthScopeMCPChallengeRead,
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		SessionVersion:      7,
		IssuedAt:            time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC),
		ExpiresAt:           time.Date(2026, 7, 1, 8, 5, 0, 0, time.UTC),
	}

	if err := store.StoreAuthorizationCode(ctx, "plain-code", claims, 5*time.Minute); err != nil {
		t.Fatalf("StoreAuthorizationCode() error = %v", err)
	}
	if keys := mini.Keys(); containsString(keys, "plain-code") {
		t.Fatalf("oauth store must hash authorization code in redis key, keys=%v", keys)
	}

	loaded, err := store.ConsumeAuthorizationCode(ctx, "plain-code")
	if err != nil {
		t.Fatalf("ConsumeAuthorizationCode() error = %v", err)
	}
	if loaded == nil || loaded.UserID != 42 || loaded.SessionVersion != 7 || loaded.CodeChallengeMethod != "S256" {
		t.Fatalf("unexpected authorization code claims: %+v", loaded)
	}
	again, err := store.ConsumeAuthorizationCode(ctx, "plain-code")
	if err != nil {
		t.Fatalf("ConsumeAuthorizationCode() second error = %v", err)
	}
	if again != nil {
		t.Fatalf("authorization code should be single use, got %+v", again)
	}
}

func TestOAuthStoreAccessAndRefreshTokensUseTTLAndRotation(t *testing.T) {
	ctx := context.Background()
	store, mini := newOAuthStoreTestEnv(t)
	claims := authcontracts.OAuthTokenClaims{
		UserID:         42,
		Username:       "alice",
		Role:           "student",
		ClientID:       "client_123",
		Scope:          authcontracts.OAuthScopeMCPChallengeRead,
		SessionVersion: 7,
		IssuedAt:       time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC),
		ExpiresAt:      time.Date(2026, 7, 1, 8, 15, 0, 0, time.UTC),
	}

	if err := store.StoreAccessToken(ctx, "access-token", claims, 15*time.Minute); err != nil {
		t.Fatalf("StoreAccessToken() error = %v", err)
	}
	loadedAccess, err := store.ResolveAccessToken(ctx, "access-token")
	if err != nil {
		t.Fatalf("ResolveAccessToken() error = %v", err)
	}
	if loadedAccess == nil || loadedAccess.ClientID != "client_123" || loadedAccess.Scope != authcontracts.OAuthScopeMCPChallengeRead {
		t.Fatalf("unexpected access token claims: %+v", loadedAccess)
	}
	mini.FastForward(16 * time.Minute)
	expiredAccess, err := store.ResolveAccessToken(ctx, "access-token")
	if err != nil {
		t.Fatalf("ResolveAccessToken() after expiry error = %v", err)
	}
	if expiredAccess != nil {
		t.Fatalf("expected expired access token to be nil, got %+v", expiredAccess)
	}

	if err := store.StoreRefreshToken(ctx, "refresh-old", claims, 30*24*time.Hour); err != nil {
		t.Fatalf("StoreRefreshToken() error = %v", err)
	}
	rotated, err := store.ConsumeRefreshToken(ctx, "refresh-old")
	if err != nil {
		t.Fatalf("ConsumeRefreshToken() error = %v", err)
	}
	if rotated == nil || rotated.UserID != 42 {
		t.Fatalf("unexpected refresh claims: %+v", rotated)
	}
	again, err := store.ConsumeRefreshToken(ctx, "refresh-old")
	if err != nil {
		t.Fatalf("ConsumeRefreshToken() second error = %v", err)
	}
	if again != nil {
		t.Fatalf("old refresh token should be consumed during rotation, got %+v", again)
	}
}

func newOAuthStoreTestEnv(t *testing.T) (*OAuthStore, *miniredis.Miniredis) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/oauth-store.sqlite"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&oauthClientRecord{}, &oauthConsentRecord{}); err != nil {
		t.Fatalf("migrate oauth store records: %v", err)
	}
	mini := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return NewOAuthStore(db, client, config.AuthOAuthConfig{RedisKeyPrefix: "test:oauth"}), mini
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
