package queries

import (
	"context"
	"strings"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

type OAuthMetadataService interface {
	ProtectedResource(ctx context.Context, requestOrigin string) (*OAuthProtectedResourceMetadataResp, error)
	AuthorizationServer(ctx context.Context, requestOrigin string) (*OAuthAuthorizationServerMetadataResp, error)
}

type OAuthProtectedResourceMetadataResp struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
}

type OAuthAuthorizationServerMetadataResp struct {
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

type oauthMetadataService struct {
	appEnv string
	config config.AuthOAuthConfig
}

func NewOAuthMetadataService(appEnv string, cfg config.AuthOAuthConfig) OAuthMetadataService {
	return &oauthMetadataService{
		appEnv: strings.TrimSpace(appEnv),
		config: cfg,
	}
}

func (s *oauthMetadataService) ProtectedResource(_ context.Context, requestOrigin string) (*OAuthProtectedResourceMetadataResp, error) {
	issuer := s.issuer(requestOrigin)
	return &OAuthProtectedResourceMetadataResp{
		Resource:             issuer + "/mcp",
		AuthorizationServers: []string{issuer},
	}, nil
}

func (s *oauthMetadataService) AuthorizationServer(_ context.Context, requestOrigin string) (*OAuthAuthorizationServerMetadataResp, error) {
	issuer := s.issuer(requestOrigin)
	return &OAuthAuthorizationServerMetadataResp{
		Issuer:                            issuer,
		AuthorizationEndpoint:             issuer + "/api/v1/oauth/authorize",
		TokenEndpoint:                     issuer + "/api/v1/oauth/token",
		RegistrationEndpoint:              issuer + "/api/v1/oauth/register",
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		TokenEndpointAuthMethodsSupported: []string{"none"},
		ScopesSupported:                   []string{authcontracts.OAuthScopeMCPChallengeRead},
	}, nil
}

func (s *oauthMetadataService) issuer(requestOrigin string) string {
	if configured := strings.TrimRight(strings.TrimSpace(s.config.IssuerURL), "/"); configured != "" {
		return configured
	}
	return strings.TrimRight(strings.TrimSpace(requestOrigin), "/")
}
