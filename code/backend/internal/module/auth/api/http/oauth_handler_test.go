package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

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

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
