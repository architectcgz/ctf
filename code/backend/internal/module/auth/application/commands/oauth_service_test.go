package commands

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

func TestOAuthServiceRegisterClientIssuesUnpredictablePublicClient(t *testing.T) {
	ctx := context.Background()
	store := &captureOAuthClientStore{}
	service := NewOAuthService(oauthServiceTestConfig(), store, zap.NewNop())

	first, err := service.RegisterClient(ctx, OAuthClientRegistrationInput{
		ClientName:              "local-codex",
		RedirectURIs:            []string{"http://127.0.0.1:14567/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   authcontracts.OAuthScopeMCPChallengeRead,
		TokenEndpointAuthMethod: "none",
	})
	if err != nil {
		t.Fatalf("RegisterClient() first error = %v", err)
	}
	second, err := service.RegisterClient(ctx, OAuthClientRegistrationInput{
		ClientName:              "local-codex",
		RedirectURIs:            []string{"http://127.0.0.1:14568/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   authcontracts.OAuthScopeMCPChallengeRead,
		TokenEndpointAuthMethod: "none",
	})
	if err != nil {
		t.Fatalf("RegisterClient() second error = %v", err)
	}

	if first.ClientID == "" || second.ClientID == "" || first.ClientID == second.ClientID {
		t.Fatalf("client ids must be non-empty and unique, first=%q second=%q", first.ClientID, second.ClientID)
	}
	if strings.Contains(first.ClientID, "local-codex") {
		t.Fatalf("client id must not be derived from client name, got %q", first.ClientID)
	}
	if first.ClientSecret != "" {
		t.Fatalf("public client registration must not return client_secret, got %q", first.ClientSecret)
	}
	if len(store.saved) != 2 || store.saved[0].TokenEndpointAuthMethod != "none" {
		t.Fatalf("expected public clients to be stored, saved=%+v", store.saved)
	}
}

func TestOAuthServiceRegisterClientRejectsUnsupportedScopeAndRedirectURI(t *testing.T) {
	service := NewOAuthService(oauthServiceTestConfig(), &captureOAuthClientStore{}, zap.NewNop())

	if _, err := service.RegisterClient(context.Background(), OAuthClientRegistrationInput{
		ClientName:              "bad-scope",
		RedirectURIs:            []string{"http://127.0.0.1:14567/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		Scope:                   "mcp:challenge:read flag:submit",
		TokenEndpointAuthMethod: "none",
	}); !isOAuthErrorCode(err, "invalid_scope") {
		t.Fatalf("expected invalid_scope for unsupported scope, got %v", err)
	}

	if _, err := service.RegisterClient(context.Background(), OAuthClientRegistrationInput{
		ClientName:              "bad-redirect",
		RedirectURIs:            []string{"https://evil.example.test/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		Scope:                   authcontracts.OAuthScopeMCPChallengeRead,
		TokenEndpointAuthMethod: "none",
	}); !isOAuthErrorCode(err, "invalid_client_metadata") {
		t.Fatalf("expected invalid_client_metadata for non-loopback redirect, got %v", err)
	}
}

func oauthServiceTestConfig() config.AuthOAuthConfig {
	return config.AuthOAuthConfig{
		AuthorizationCodeTTL:      5 * time.Minute,
		AccessTokenTTL:            15 * time.Minute,
		RefreshTokenTTL:           30 * 24 * time.Hour,
		ClientRegistrationEnabled: true,
		RedisKeyPrefix:            "test:oauth",
	}
}

type captureOAuthClientStore struct {
	saved []authcontracts.OAuthClient
}

func (s *captureOAuthClientStore) SaveClient(ctx context.Context, client authcontracts.OAuthClient) error {
	s.saved = append(s.saved, client)
	return nil
}

func isOAuthErrorCode(err error, code string) bool {
	oauthErr, ok := err.(*authcontracts.OAuthError)
	return ok && oauthErr.Code == code
}
