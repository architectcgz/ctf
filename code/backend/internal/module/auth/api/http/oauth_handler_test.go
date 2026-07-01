package http_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

type testOAuthAuthorizationServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
}

type testOAuthProtectedResourceMetadata struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
}

type testOAuthClientRegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

type testOAuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type testOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

func TestHTTP_OAuthWellKnownMetadata(t *testing.T) {
	env := newIntegrationTestEnv(t)

	protectedResp := performJSONRequest(t, env.router, http.MethodGet, "https://ctf.example.test/.well-known/oauth-protected-resource", nil, nil, nil)
	if protectedResp.Code != http.StatusOK {
		t.Fatalf("unexpected protected resource metadata status: %d body=%s", protectedResp.Code, protectedResp.Body.String())
	}
	protected := decodeJSON[testOAuthProtectedResourceMetadata](t, protectedResp.Body.Bytes())
	if protected.Resource != "https://ctf.example.test/mcp" {
		t.Fatalf("resource = %q, want https://ctf.example.test/mcp", protected.Resource)
	}
	if len(protected.AuthorizationServers) != 1 || protected.AuthorizationServers[0] != "https://ctf.example.test" {
		t.Fatalf("authorization_servers = %+v", protected.AuthorizationServers)
	}

	authResp := performJSONRequest(t, env.router, http.MethodGet, "https://ctf.example.test/.well-known/oauth-authorization-server", nil, nil, nil)
	if authResp.Code != http.StatusOK {
		t.Fatalf("unexpected authorization server metadata status: %d body=%s", authResp.Code, authResp.Body.String())
	}
	meta := decodeJSON[testOAuthAuthorizationServerMetadata](t, authResp.Body.Bytes())
	if meta.Issuer != "https://ctf.example.test" {
		t.Fatalf("issuer = %q, want https://ctf.example.test", meta.Issuer)
	}
	if meta.AuthorizationEndpoint != "https://ctf.example.test/api/v1/oauth/authorize" {
		t.Fatalf("authorization endpoint = %q", meta.AuthorizationEndpoint)
	}
	if !containsString(meta.CodeChallengeMethodsSupported, "S256") {
		t.Fatalf("code_challenge_methods_supported = %+v", meta.CodeChallengeMethodsSupported)
	}
	if !containsString(meta.TokenEndpointAuthMethodsSupported, "none") {
		t.Fatalf("token_endpoint_auth_methods_supported = %+v", meta.TokenEndpointAuthMethodsSupported)
	}
	if !containsString(meta.ScopesSupported, authcontracts.OAuthScopeMCPChallengeRead) {
		t.Fatalf("scopes_supported = %+v", meta.ScopesSupported)
	}
}

func TestHTTP_OAuthClientRegistrationPublicLoopback(t *testing.T) {
	env := newIntegrationTestEnv(t)

	resp := performJSONRequest(t, env.router, http.MethodPost, "/api/v1/oauth/register", map[string]any{
		"client_name":                "local-codex",
		"redirect_uris":              []string{"http://127.0.0.1:14567/callback"},
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"scope":                      authcontracts.OAuthScopeMCPChallengeRead,
		"token_endpoint_auth_method": "none",
	}, nil, nil)
	if resp.Code != http.StatusCreated {
		t.Fatalf("unexpected register status: %d body=%s", resp.Code, resp.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resp.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw registration response: %v body=%s", err, resp.Body.String())
	}
	if _, exists := raw["client_secret"]; exists {
		t.Fatalf("public client registration must not include client_secret, body=%s", resp.Body.String())
	}
	body := decodeJSON[testOAuthClientRegistrationResponse](t, resp.Body.Bytes())
	if body.ClientID == "" || body.TokenEndpointAuthMethod != "none" {
		t.Fatalf("unexpected registration response: %+v", body)
	}
}

func TestHTTP_OAuthClientRegistrationRejectsNonLoopbackByDefault(t *testing.T) {
	env := newIntegrationTestEnv(t)

	resp := performJSONRequest(t, env.router, http.MethodPost, "/api/v1/oauth/register", map[string]any{
		"client_name":                "remote-agent",
		"redirect_uris":              []string{"https://agent.example.test/callback"},
		"grant_types":                []string{"authorization_code"},
		"response_types":             []string{"code"},
		"scope":                      authcontracts.OAuthScopeMCPChallengeRead,
		"token_endpoint_auth_method": "none",
	}, nil, nil)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unexpected invalid register status: %d body=%s", resp.Code, resp.Body.String())
	}
	body := decodeJSON[testOAuthErrorResponse](t, resp.Body.Bytes())
	if body.Error != "invalid_client_metadata" {
		t.Fatalf("oauth error = %+v", body)
	}
}

func TestHTTP_OAuthAuthorizeValidatesRequest(t *testing.T) {
	env := newIntegrationTestEnv(t)
	registerOAuthTestClient(t, env)

	testCases := []struct {
		name   string
		target string
	}{
		{
			name:   "response_type",
			target: "/api/v1/oauth/authorize?response_type=token&client_id=client_test&redirect_uri=http%3A%2F%2F127.0.0.1%3A14567%2Fcallback&scope=mcp%3Achallenge%3Aread&code_challenge=challenge&code_challenge_method=S256",
		},
		{
			name:   "redirect_uri",
			target: "/api/v1/oauth/authorize?response_type=code&client_id=client_test&redirect_uri=http%3A%2F%2F127.0.0.1%3A14567%2Fcallback%2Fextra&scope=mcp%3Achallenge%3Aread&code_challenge=challenge&code_challenge_method=S256",
		},
		{
			name:   "code_challenge",
			target: "/api/v1/oauth/authorize?response_type=code&client_id=client_test&redirect_uri=http%3A%2F%2F127.0.0.1%3A14567%2Fcallback&scope=mcp%3Achallenge%3Aread&code_challenge_method=S256",
		},
		{
			name:   "code_challenge_method",
			target: "/api/v1/oauth/authorize?response_type=code&client_id=client_test&redirect_uri=http%3A%2F%2F127.0.0.1%3A14567%2Fcallback&scope=mcp%3Achallenge%3Aread&code_challenge=challenge&code_challenge_method=plain",
		},
		{
			name:   "scope",
			target: "/api/v1/oauth/authorize?response_type=code&client_id=client_test&redirect_uri=http%3A%2F%2F127.0.0.1%3A14567%2Fcallback&scope=flag%3Asubmit&code_challenge=challenge&code_challenge_method=S256",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := performJSONRequest(t, env.router, http.MethodGet, tc.target, nil, nil, nil)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("expected invalid authorize request to return 400, got %d body=%s", resp.Code, resp.Body.String())
			}
			body := decodeJSON[testOAuthErrorResponse](t, resp.Body.Bytes())
			if body.Error == "" {
				t.Fatalf("expected oauth error body, got %+v", body)
			}
		})
	}
}

func TestHTTP_OAuthAuthorizeRedirectsUnauthenticatedUserToLogin(t *testing.T) {
	env := newIntegrationTestEnvWithAuthConfig(t, func(cfg *config.AuthConfig) {
		cfg.OAuth.LoginURL = "http://127.0.0.1:5174/login"
	})
	registerOAuthTestClient(t, env)

	target := validAuthorizeTarget("state-login")
	resp := performJSONRequest(t, env.router, http.MethodGet, target, nil, nil, nil)
	if resp.Code != http.StatusFound {
		t.Fatalf("expected unauthenticated authorize to redirect, got %d body=%s", resp.Code, resp.Body.String())
	}
	location := resp.Header().Get("Location")
	if !strings.HasPrefix(location, "http://127.0.0.1:5174/login?redirect=") {
		t.Fatalf("unexpected login redirect location: %q", location)
	}
	redirectValue, err := url.QueryUnescape(strings.TrimPrefix(location, "http://127.0.0.1:5174/login?redirect="))
	if err != nil {
		t.Fatalf("decode redirect: %v", err)
	}
	if redirectValue != target {
		t.Fatalf("login redirect target = %q, want %q", redirectValue, target)
	}
}

func TestHTTP_OAuthAuthorizeShowsConsentForLoggedInUser(t *testing.T) {
	env := newIntegrationTestEnv(t)
	registerOAuthTestClient(t, env)
	sessionCookie := loginOAuthTestUser(t, env)

	resp := performJSONRequest(t, env.router, http.MethodGet, validAuthorizeTarget("state-consent"), nil, nil, []*http.Cookie{sessionCookie})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected consent page, got %d body=%s", resp.Code, resp.Body.String())
	}
	contentType := resp.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("consent content type = %q", contentType)
	}
	body := resp.Body.String()
	for _, required := range []string{
		`action="/api/v1/oauth/authorize"`,
		`name="csrf_nonce"`,
		`name="client_id" value="client_test"`,
		`local-codex`,
		authcontracts.OAuthScopeMCPChallengeRead,
		"oauth_user",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("consent page missing %q:\n%s", required, body)
		}
	}
	if strings.Contains(body, "access_token") || strings.Contains(body, "refresh_token") {
		t.Fatalf("consent page must not expose tokens:\n%s", body)
	}
}

func TestHTTP_OAuthAuthorizeApproveAndDenyRedirect(t *testing.T) {
	env := newIntegrationTestEnv(t)
	registerOAuthTestClient(t, env)
	sessionCookie := loginOAuthTestUser(t, env)

	consentResp := performJSONRequest(t, env.router, http.MethodGet, validAuthorizeTarget("state-approve"), nil, nil, []*http.Cookie{sessionCookie})
	nonce := extractConsentNonce(t, consentResp.Body.String())
	approveResp := postOAuthAuthorizeForm(t, env, sessionCookie, nonce, "true", "state-approve")
	if approveResp.Code != http.StatusFound {
		t.Fatalf("expected approve redirect, got %d body=%s", approveResp.Code, approveResp.Body.String())
	}
	approveLocation := approveResp.Header().Get("Location")
	if !strings.HasPrefix(approveLocation, "http://127.0.0.1:14567/callback?") ||
		!strings.Contains(approveLocation, "code=") ||
		!strings.Contains(approveLocation, "state=state-approve") {
		t.Fatalf("unexpected approve redirect: %q", approveLocation)
	}

	denyEnv := newIntegrationTestEnv(t)
	registerOAuthTestClient(t, denyEnv)
	denySessionCookie := loginOAuthTestUser(t, denyEnv)
	consentResp = performJSONRequest(t, denyEnv.router, http.MethodGet, validAuthorizeTarget("state-deny"), nil, nil, []*http.Cookie{denySessionCookie})
	nonce = extractConsentNonce(t, consentResp.Body.String())
	denyResp := postOAuthAuthorizeForm(t, denyEnv, denySessionCookie, nonce, "false", "state-deny")
	if denyResp.Code != http.StatusFound {
		t.Fatalf("expected deny redirect, got %d body=%s", denyResp.Code, denyResp.Body.String())
	}
	denyLocation := denyResp.Header().Get("Location")
	if !strings.HasPrefix(denyLocation, "http://127.0.0.1:14567/callback?") ||
		!strings.Contains(denyLocation, "error=access_denied") ||
		!strings.Contains(denyLocation, "state=state-deny") {
		t.Fatalf("unexpected deny redirect: %q", denyLocation)
	}
}

func TestHTTP_OAuthTokenAuthorizationCodeGrantIssuesTokens(t *testing.T) {
	env := newIntegrationTestEnv(t)
	registerOAuthTestClient(t, env)
	sessionCookie := loginOAuthTestUser(t, env)
	code := approveOAuthTestAuthorizationCode(t, env, sessionCookie, "state-token")

	resp := postOAuthTokenForm(t, env, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {"client_test"},
		"code":          {code},
		"redirect_uri":  {"http://127.0.0.1:14567/callback"},
		"code_verifier": {oauthHandlerTestPKCEVerifier},
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected token response, got %d body=%s", resp.Code, resp.Body.String())
	}
	body := decodeJSON[testOAuthTokenResponse](t, resp.Body.Bytes())
	if body.AccessToken == "" || body.RefreshToken == "" || body.AccessToken == body.RefreshToken {
		t.Fatalf("expected distinct access and refresh tokens, got %+v", body)
	}
	if body.TokenType != "Bearer" || body.ExpiresIn <= 0 || body.Scope != authcontracts.OAuthScopeMCPChallengeRead {
		t.Fatalf("unexpected token response metadata: %+v", body)
	}
}

func TestHTTP_OAuthTokenRejectsPKCEAndRedirectMismatch(t *testing.T) {
	testCases := []struct {
		name string
		form func(code string) url.Values
	}{
		{
			name: "missing verifier",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type":   {"authorization_code"},
					"client_id":    {"client_test"},
					"code":         {code},
					"redirect_uri": {"http://127.0.0.1:14567/callback"},
				}
			},
		},
		{
			name: "wrong verifier",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type":    {"authorization_code"},
					"client_id":     {"client_test"},
					"code":          {code},
					"redirect_uri":  {"http://127.0.0.1:14567/callback"},
					"code_verifier": {"wrong-verifier"},
				}
			},
		},
		{
			name: "redirect mismatch",
			form: func(code string) url.Values {
				return url.Values{
					"grant_type":    {"authorization_code"},
					"client_id":     {"client_test"},
					"code":          {code},
					"redirect_uri":  {"http://127.0.0.1:14567/other"},
					"code_verifier": {oauthHandlerTestPKCEVerifier},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newIntegrationTestEnv(t)
			registerOAuthTestClient(t, env)
			sessionCookie := loginOAuthTestUser(t, env)
			code := approveOAuthTestAuthorizationCode(t, env, sessionCookie, "state-"+tc.name)

			resp := postOAuthTokenForm(t, env, tc.form(code))
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("expected token error, got %d body=%s", resp.Code, resp.Body.String())
			}
			body := decodeJSON[testOAuthErrorResponse](t, resp.Body.Bytes())
			if body.Error == "" {
				t.Fatalf("expected oauth error body, got %+v", body)
			}
		})
	}
}

func TestHTTP_OAuthRefreshTokenRotatesAndRejectsReuse(t *testing.T) {
	env := newIntegrationTestEnv(t)
	registerOAuthTestClient(t, env)
	sessionCookie := loginOAuthTestUser(t, env)
	code := approveOAuthTestAuthorizationCode(t, env, sessionCookie, "state-refresh")

	tokenResp := postOAuthTokenForm(t, env, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {"client_test"},
		"code":          {code},
		"redirect_uri":  {"http://127.0.0.1:14567/callback"},
		"code_verifier": {oauthHandlerTestPKCEVerifier},
	})
	first := decodeJSON[testOAuthTokenResponse](t, tokenResp.Body.Bytes())

	refreshResp := postOAuthTokenForm(t, env, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"client_test"},
		"refresh_token": {first.RefreshToken},
	})
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("expected refresh success, got %d body=%s", refreshResp.Code, refreshResp.Body.String())
	}
	second := decodeJSON[testOAuthTokenResponse](t, refreshResp.Body.Bytes())
	if second.AccessToken == "" || second.RefreshToken == "" || second.RefreshToken == first.RefreshToken {
		t.Fatalf("expected rotated refresh token, first=%+v second=%+v", first, second)
	}

	reuseResp := postOAuthTokenForm(t, env, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"client_test"},
		"refresh_token": {first.RefreshToken},
	})
	if reuseResp.Code != http.StatusBadRequest {
		t.Fatalf("expected refresh reuse failure, got %d body=%s", reuseResp.Code, reuseResp.Body.String())
	}
	reuseErr := decodeJSON[testOAuthErrorResponse](t, reuseResp.Body.Bytes())
	if reuseErr.Error != "invalid_grant" {
		t.Fatalf("expected invalid_grant for reused refresh token, got %+v", reuseErr)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func registerOAuthTestClient(t *testing.T, env *integrationTestEnv) {
	t.Helper()
	env.oauthStore.clients["client_test"] = authcontracts.OAuthClient{
		ClientID:                "client_test",
		ClientName:              "local-codex",
		RedirectURIs:            []string{"http://127.0.0.1:14567/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   authcontracts.OAuthScopeMCPChallengeRead,
		TokenEndpointAuthMethod: "none",
	}
}

func loginOAuthTestUser(t *testing.T, env *integrationTestEnv) *http.Cookie {
	t.Helper()
	resp := performJSONRequest(t, env.router, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": "oauth_user",
		"password": "Password123",
	}, nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("register oauth user: status=%d body=%s", resp.Code, resp.Body.String())
	}
	return cloneCookie(resp.Result().Cookies(), "ctf_session")
}

func validAuthorizeTarget(state string) string {
	return "/api/v1/oauth/authorize?response_type=code&client_id=client_test&redirect_uri=http%3A%2F%2F127.0.0.1%3A14567%2Fcallback&scope=mcp%3Achallenge%3Aread&state=" + url.QueryEscape(state) + "&code_challenge=" + oauthHandlerTestPKCEChallenge() + "&code_challenge_method=S256"
}

func extractConsentNonce(t *testing.T, body string) string {
	t.Helper()
	matches := regexp.MustCompile(`name="csrf_nonce" value="([^"]+)"`).FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("csrf nonce missing in consent body:\n%s", body)
	}
	return matches[1]
}

func postOAuthAuthorizeForm(t *testing.T, env *integrationTestEnv, sessionCookie *http.Cookie, nonce, approve, state string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{}
	form.Set("response_type", "code")
	form.Set("client_id", "client_test")
	form.Set("redirect_uri", "http://127.0.0.1:14567/callback")
	form.Set("scope", authcontracts.OAuthScopeMCPChallengeRead)
	form.Set("state", state)
	form.Set("code_challenge", oauthHandlerTestPKCEChallenge())
	form.Set("code_challenge_method", "S256")
	form.Set("csrf_nonce", nonce)
	form.Set("approve", approve)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec
}

const oauthHandlerTestPKCEVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

func oauthHandlerTestPKCEChallenge() string {
	sum := sha256.Sum256([]byte(oauthHandlerTestPKCEVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func approveOAuthTestAuthorizationCode(t *testing.T, env *integrationTestEnv, sessionCookie *http.Cookie, state string) string {
	t.Helper()
	consentResp := performJSONRequest(t, env.router, http.MethodGet, validAuthorizeTarget(state), nil, nil, []*http.Cookie{sessionCookie})
	if consentResp.Code != http.StatusOK {
		t.Fatalf("expected consent page, got %d body=%s", consentResp.Code, consentResp.Body.String())
	}
	nonce := extractConsentNonce(t, consentResp.Body.String())
	approveResp := postOAuthAuthorizeForm(t, env, sessionCookie, nonce, "true", state)
	if approveResp.Code != http.StatusFound {
		t.Fatalf("expected approve redirect, got %d body=%s", approveResp.Code, approveResp.Body.String())
	}
	location := approveResp.Header().Get("Location")
	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse approve redirect %q: %v", location, err)
	}
	code := parsed.Query().Get("code")
	if code == "" {
		t.Fatalf("approve redirect missing code: %q", location)
	}
	return code
}

func postOAuthTokenForm(t *testing.T, env *integrationTestEnv, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec
}
