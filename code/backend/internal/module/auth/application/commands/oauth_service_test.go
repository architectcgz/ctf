package commands

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

func TestOAuthServiceRegisterClientIssuesUnpredictablePublicClient(t *testing.T) {
	ctx := context.Background()
	store := newCaptureOAuthClientStore()
	service := NewOAuthService(oauthServiceTestConfig(), store, newCaptureSessionVersionReader(), zap.NewNop())

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
	service := NewOAuthService(oauthServiceTestConfig(), newCaptureOAuthClientStore(), newCaptureSessionVersionReader(), zap.NewNop())

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

func TestOAuthServiceExchangeAuthorizationCodeIssuesBearerTokens(t *testing.T) {
	ctx := context.Background()
	store := newCaptureOAuthClientStore()
	store.clients["client_test"] = oauthServiceTestClient()
	store.codes["code_valid"] = oauthServiceTestAuthorizationCode("client_test", oauthServiceTestPKCEChallenge(), 7)
	versions := newCaptureSessionVersionReader()
	versions.versions[42] = 7
	service := NewOAuthService(oauthServiceTestConfig(), store, versions, zap.NewNop())

	resp, err := service.ExchangeAuthorizationCode(ctx, OAuthAuthorizationCodeExchangeInput{
		ClientID:     "client_test",
		Code:         "code_valid",
		RedirectURI:  "http://127.0.0.1:14567/callback",
		CodeVerifier: oauthServiceTestPKCEVerifier,
	})
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode() error = %v", err)
	}

	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.AccessToken == resp.RefreshToken {
		t.Fatalf("expected distinct access and refresh tokens, got %+v", resp)
	}
	if resp.TokenType != "Bearer" || resp.Scope != authcontracts.OAuthScopeMCPChallengeRead {
		t.Fatalf("unexpected token response metadata: %+v", resp)
	}
	if resp.ExpiresIn <= 0 || resp.ExpiresIn > int64(oauthServiceTestConfig().AccessTokenTTL.Seconds()) {
		t.Fatalf("expires_in = %d", resp.ExpiresIn)
	}
	if _, exists := store.accessTokens[resp.AccessToken]; !exists {
		t.Fatalf("access token was not stored: %+v", store.accessTokens)
	}
	if _, exists := store.refreshTokens[resp.RefreshToken]; !exists {
		t.Fatalf("refresh token was not stored: %+v", store.refreshTokens)
	}
	if _, exists := store.codes["code_valid"]; exists {
		t.Fatal("authorization code must be consumed after token exchange")
	}
}

func TestOAuthServiceExchangeAuthorizationCodeRejectsPKCEAndRedirectMismatch(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name         string
		code         authcontracts.OAuthAuthorizationCode
		input        OAuthAuthorizationCodeExchangeInput
		expectedCode string
	}{
		{
			name: "missing verifier",
			code: oauthServiceTestAuthorizationCode("client_test", oauthServiceTestPKCEChallenge(), 7),
			input: OAuthAuthorizationCodeExchangeInput{
				ClientID:    "client_test",
				Code:        "code_missing_verifier",
				RedirectURI: "http://127.0.0.1:14567/callback",
			},
			expectedCode: "invalid_request",
		},
		{
			name: "wrong verifier",
			code: oauthServiceTestAuthorizationCode("client_test", oauthServiceTestPKCEChallenge(), 7),
			input: OAuthAuthorizationCodeExchangeInput{
				ClientID:     "client_test",
				Code:         "code_wrong_verifier",
				RedirectURI:  "http://127.0.0.1:14567/callback",
				CodeVerifier: "wrong-verifier",
			},
			expectedCode: "invalid_grant",
		},
		{
			name: "plain method is never accepted",
			code: oauthServiceTestAuthorizationCodeWithMethod("client_test", "plain-challenge", "plain", 7),
			input: OAuthAuthorizationCodeExchangeInput{
				ClientID:     "client_test",
				Code:         "code_plain_method",
				RedirectURI:  "http://127.0.0.1:14567/callback",
				CodeVerifier: "plain-challenge",
			},
			expectedCode: "invalid_grant",
		},
		{
			name: "redirect uri mismatch",
			code: oauthServiceTestAuthorizationCode("client_test", oauthServiceTestPKCEChallenge(), 7),
			input: OAuthAuthorizationCodeExchangeInput{
				ClientID:     "client_test",
				Code:         "code_redirect_mismatch",
				RedirectURI:  "http://127.0.0.1:14567/other",
				CodeVerifier: oauthServiceTestPKCEVerifier,
			},
			expectedCode: "invalid_grant",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := newCaptureOAuthClientStore()
			store.clients["client_test"] = oauthServiceTestClient()
			store.codes[tc.input.Code] = tc.code
			versions := newCaptureSessionVersionReader()
			versions.versions[42] = 7
			service := NewOAuthService(oauthServiceTestConfig(), store, versions, zap.NewNop())

			if _, err := service.ExchangeAuthorizationCode(ctx, tc.input); !isOAuthErrorCode(err, tc.expectedCode) {
				t.Fatalf("expected %s, got %v", tc.expectedCode, err)
			}
		})
	}

	store := newCaptureOAuthClientStore()
	store.clients["client_test"] = oauthServiceTestClient()
	store.codes["code_replay"] = oauthServiceTestAuthorizationCode("client_test", oauthServiceTestPKCEChallenge(), 7)
	versions := newCaptureSessionVersionReader()
	versions.versions[42] = 7
	service := NewOAuthService(oauthServiceTestConfig(), store, versions, zap.NewNop())
	if _, err := service.ExchangeAuthorizationCode(ctx, OAuthAuthorizationCodeExchangeInput{
		ClientID:     "client_test",
		Code:         "code_replay",
		RedirectURI:  "http://127.0.0.1:14567/callback",
		CodeVerifier: oauthServiceTestPKCEVerifier,
	}); err != nil {
		t.Fatalf("first exchange error = %v", err)
	}
	if _, err := service.ExchangeAuthorizationCode(ctx, OAuthAuthorizationCodeExchangeInput{
		ClientID:     "client_test",
		Code:         "code_replay",
		RedirectURI:  "http://127.0.0.1:14567/callback",
		CodeVerifier: oauthServiceTestPKCEVerifier,
	}); !isOAuthErrorCode(err, "invalid_grant") {
		t.Fatalf("expected invalid_grant for replayed code, got %v", err)
	}
}

func TestOAuthServiceRefreshRotatesTokenAndRejectsStaleSessionVersion(t *testing.T) {
	ctx := context.Background()
	store := newCaptureOAuthClientStore()
	store.clients["client_test"] = oauthServiceTestClient()
	store.refreshTokens["refresh_old"] = oauthServiceTestTokenClaims("client_test", 7)
	versions := newCaptureSessionVersionReader()
	versions.versions[42] = 7
	service := NewOAuthService(oauthServiceTestConfig(), store, versions, zap.NewNop())

	resp, err := service.RefreshAccessToken(ctx, OAuthRefreshTokenInput{
		ClientID:     "client_test",
		RefreshToken: "refresh_old",
	})
	if err != nil {
		t.Fatalf("RefreshAccessToken() error = %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.RefreshToken == "refresh_old" {
		t.Fatalf("expected rotated refresh token and access token, got %+v", resp)
	}
	if _, exists := store.refreshTokens["refresh_old"]; exists {
		t.Fatal("old refresh token must be consumed during rotation")
	}
	if _, err := service.RefreshAccessToken(ctx, OAuthRefreshTokenInput{
		ClientID:     "client_test",
		RefreshToken: "refresh_old",
	}); !isOAuthErrorCode(err, "invalid_grant") {
		t.Fatalf("expected invalid_grant for reused refresh token, got %v", err)
	}

	store.refreshTokens["refresh_stale"] = oauthServiceTestTokenClaims("client_test", 7)
	store.accessTokens["access_stale"] = oauthServiceTestTokenClaims("client_test", 7)
	versions.versions[42] = 8
	if _, err := service.RefreshAccessToken(ctx, OAuthRefreshTokenInput{
		ClientID:     "client_test",
		RefreshToken: "refresh_stale",
	}); !isOAuthErrorCode(err, "invalid_grant") {
		t.Fatalf("expected invalid_grant for stale refresh session version, got %v", err)
	}
	if _, err := service.ResolveOAuthAccessToken(ctx, "access_stale", authcontracts.OAuthScopeMCPChallengeRead); !isOAuthErrorCode(err, "invalid_grant") {
		t.Fatalf("expected invalid_grant for stale access session version, got %v", err)
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

const oauthServiceTestPKCEVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

func oauthServiceTestPKCEChallenge() string {
	sum := sha256.Sum256([]byte(oauthServiceTestPKCEVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func oauthServiceTestClient() authcontracts.OAuthClient {
	return authcontracts.OAuthClient{
		ClientID:                "client_test",
		ClientName:              "local-codex",
		RedirectURIs:            []string{"http://127.0.0.1:14567/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   authcontracts.OAuthScopeMCPChallengeRead,
		TokenEndpointAuthMethod: "none",
	}
}

func oauthServiceTestAuthorizationCode(clientID, challenge string, sessionVersion int64) authcontracts.OAuthAuthorizationCode {
	return oauthServiceTestAuthorizationCodeWithMethod(clientID, challenge, "S256", sessionVersion)
}

func oauthServiceTestAuthorizationCodeWithMethod(clientID, challenge, method string, sessionVersion int64) authcontracts.OAuthAuthorizationCode {
	now := time.Now().UTC()
	return authcontracts.OAuthAuthorizationCode{
		UserID:              42,
		Username:            "alice",
		Role:                "student",
		ClientID:            clientID,
		RedirectURI:         "http://127.0.0.1:14567/callback",
		Scope:               authcontracts.OAuthScopeMCPChallengeRead,
		CodeChallenge:       challenge,
		CodeChallengeMethod: method,
		SessionVersion:      sessionVersion,
		IssuedAt:            now,
		ExpiresAt:           now.Add(5 * time.Minute),
	}
}

func oauthServiceTestTokenClaims(clientID string, sessionVersion int64) authcontracts.OAuthTokenClaims {
	now := time.Now().UTC()
	return authcontracts.OAuthTokenClaims{
		UserID:         42,
		Username:       "alice",
		Role:           "student",
		ClientID:       clientID,
		Scope:          authcontracts.OAuthScopeMCPChallengeRead,
		SessionVersion: sessionVersion,
		IssuedAt:       now,
		ExpiresAt:      now.Add(30 * time.Minute),
	}
}

type captureOAuthClientStore struct {
	saved         []authcontracts.OAuthClient
	clients       map[string]authcontracts.OAuthClient
	consents      map[string]authcontracts.OAuthConsent
	codes         map[string]authcontracts.OAuthAuthorizationCode
	accessTokens  map[string]authcontracts.OAuthTokenClaims
	refreshTokens map[string]authcontracts.OAuthTokenClaims
	nonces        map[string]bool
}

func newCaptureOAuthClientStore() *captureOAuthClientStore {
	return &captureOAuthClientStore{
		clients:       make(map[string]authcontracts.OAuthClient),
		consents:      make(map[string]authcontracts.OAuthConsent),
		codes:         make(map[string]authcontracts.OAuthAuthorizationCode),
		accessTokens:  make(map[string]authcontracts.OAuthTokenClaims),
		refreshTokens: make(map[string]authcontracts.OAuthTokenClaims),
		nonces:        make(map[string]bool),
	}
}

func (s *captureOAuthClientStore) SaveClient(ctx context.Context, client authcontracts.OAuthClient) error {
	s.saved = append(s.saved, client)
	s.clients[client.ClientID] = client
	return nil
}

func (s *captureOAuthClientStore) FindClientByID(ctx context.Context, clientID string) (*authcontracts.OAuthClient, error) {
	client, exists := s.clients[clientID]
	if !exists {
		return nil, nil
	}
	return &client, nil
}

func (s *captureOAuthClientStore) FindActiveConsent(ctx context.Context, userID int64, clientID, scope string) (*authcontracts.OAuthConsent, error) {
	return nil, nil
}

func (s *captureOAuthClientStore) UpsertConsent(ctx context.Context, consent authcontracts.OAuthConsent) error {
	return nil
}

func (s *captureOAuthClientStore) StoreAuthorizationCode(ctx context.Context, code string, claims authcontracts.OAuthAuthorizationCode, ttl time.Duration) error {
	s.codes[code] = claims
	return nil
}

func (s *captureOAuthClientStore) ConsumeAuthorizationCode(ctx context.Context, code string) (*authcontracts.OAuthAuthorizationCode, error) {
	claims, exists := s.codes[code]
	if !exists {
		return nil, nil
	}
	delete(s.codes, code)
	return &claims, nil
}

func (s *captureOAuthClientStore) StoreAccessToken(ctx context.Context, token string, claims authcontracts.OAuthTokenClaims, ttl time.Duration) error {
	s.accessTokens[token] = claims
	return nil
}

func (s *captureOAuthClientStore) ResolveAccessToken(ctx context.Context, token string) (*authcontracts.OAuthTokenClaims, error) {
	claims, exists := s.accessTokens[token]
	if !exists {
		return nil, nil
	}
	return &claims, nil
}

func (s *captureOAuthClientStore) StoreRefreshToken(ctx context.Context, token string, claims authcontracts.OAuthTokenClaims, ttl time.Duration) error {
	s.refreshTokens[token] = claims
	return nil
}

func (s *captureOAuthClientStore) ConsumeRefreshToken(ctx context.Context, token string) (*authcontracts.OAuthTokenClaims, error) {
	claims, exists := s.refreshTokens[token]
	if !exists {
		return nil, nil
	}
	delete(s.refreshTokens, token)
	return &claims, nil
}

func (s *captureOAuthClientStore) StoreConsentNonce(ctx context.Context, nonce string, ttl time.Duration) error {
	s.nonces[nonce] = true
	return nil
}

func (s *captureOAuthClientStore) ConsumeConsentNonce(ctx context.Context, nonce string) (bool, error) {
	if !s.nonces[nonce] {
		return false, nil
	}
	delete(s.nonces, nonce)
	return true, nil
}

type captureSessionVersionReader struct {
	versions map[int64]int64
}

func newCaptureSessionVersionReader() *captureSessionVersionReader {
	return &captureSessionVersionReader{versions: make(map[int64]int64)}
}

func (r *captureSessionVersionReader) CurrentSessionVersion(ctx context.Context, userID int64) (int64, error) {
	return r.versions[userID], nil
}

func isOAuthErrorCode(err error, code string) bool {
	oauthErr, ok := err.(*authcontracts.OAuthError)
	return ok && oauthErr.Code == code
}
