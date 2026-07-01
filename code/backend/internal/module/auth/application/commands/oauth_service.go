package commands

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

type OAuthService interface {
	RegisterClient(ctx context.Context, req OAuthClientRegistrationInput) (*OAuthClientRegistrationResp, error)
}

type OAuthClientStore interface {
	SaveClient(ctx context.Context, client authcontracts.OAuthClient) error
}

type OAuthClientRegistrationInput struct {
	ClientName              string
	ClientURI               string
	RedirectURIs            []string
	GrantTypes              []string
	ResponseTypes           []string
	Scope                   string
	TokenEndpointAuthMethod string
}

type OAuthClientRegistrationResp struct {
	ClientID                string
	ClientSecret            string
	ClientName              string
	ClientURI               string
	RedirectURIs            []string
	GrantTypes              []string
	ResponseTypes           []string
	Scope                   string
	TokenEndpointAuthMethod string
	CreatedAt               time.Time
}

type oauthService struct {
	config config.AuthOAuthConfig
	store  OAuthClientStore
	log    *zap.Logger
}

func NewOAuthService(cfg config.AuthOAuthConfig, store OAuthClientStore, log *zap.Logger) OAuthService {
	if log == nil {
		log = zap.NewNop()
	}
	return &oauthService{
		config: cfg,
		store:  store,
		log:    log,
	}
}

func (s *oauthService) RegisterClient(ctx context.Context, req OAuthClientRegistrationInput) (*OAuthClientRegistrationResp, error) {
	if !s.config.ClientRegistrationEnabled {
		return nil, authcontracts.NewOAuthInvalidClientMetadata("dynamic client registration is disabled")
	}
	if s.store == nil {
		return nil, apperror.ErrServiceUnavailable
	}

	normalized, err := s.normalizeRegistration(req)
	if err != nil {
		return nil, err
	}

	clientID, err := generateOAuthClientID()
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	now := time.Now().UTC()
	client := authcontracts.OAuthClient{
		ClientID:                clientID,
		ClientName:              normalized.ClientName,
		ClientURI:               normalized.ClientURI,
		RedirectURIs:            normalized.RedirectURIs,
		GrantTypes:              normalized.GrantTypes,
		ResponseTypes:           normalized.ResponseTypes,
		Scope:                   normalized.Scope,
		TokenEndpointAuthMethod: normalized.TokenEndpointAuthMethod,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := s.store.SaveClient(ctx, client); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &OAuthClientRegistrationResp{
		ClientID:                client.ClientID,
		ClientName:              client.ClientName,
		ClientURI:               client.ClientURI,
		RedirectURIs:            append([]string(nil), client.RedirectURIs...),
		GrantTypes:              append([]string(nil), client.GrantTypes...),
		ResponseTypes:           append([]string(nil), client.ResponseTypes...),
		Scope:                   client.Scope,
		TokenEndpointAuthMethod: client.TokenEndpointAuthMethod,
		CreatedAt:               client.CreatedAt,
	}, nil
}

func (s *oauthService) normalizeRegistration(req OAuthClientRegistrationInput) (OAuthClientRegistrationInput, error) {
	req.ClientName = strings.TrimSpace(req.ClientName)
	req.ClientURI = strings.TrimSpace(req.ClientURI)
	req.TokenEndpointAuthMethod = strings.TrimSpace(req.TokenEndpointAuthMethod)
	if req.TokenEndpointAuthMethod == "" {
		req.TokenEndpointAuthMethod = "none"
	}
	if req.ClientName == "" {
		return OAuthClientRegistrationInput{}, authcontracts.NewOAuthInvalidClientMetadata("client_name is required")
	}
	if req.TokenEndpointAuthMethod != "none" {
		return OAuthClientRegistrationInput{}, authcontracts.NewOAuthInvalidClientMetadata("only public clients with token_endpoint_auth_method=none are supported")
	}

	grantTypes, err := normalizeOAuthGrantTypes(req.GrantTypes)
	if err != nil {
		return OAuthClientRegistrationInput{}, err
	}
	responseTypes, err := normalizeOAuthResponseTypes(req.ResponseTypes)
	if err != nil {
		return OAuthClientRegistrationInput{}, err
	}
	scope, err := normalizeOAuthScope(req.Scope)
	if err != nil {
		return OAuthClientRegistrationInput{}, err
	}
	redirectURIs, err := s.normalizeRedirectURIs(req.RedirectURIs)
	if err != nil {
		return OAuthClientRegistrationInput{}, err
	}

	req.GrantTypes = grantTypes
	req.ResponseTypes = responseTypes
	req.Scope = scope
	req.RedirectURIs = redirectURIs
	return req, nil
}

func normalizeOAuthGrantTypes(values []string) ([]string, error) {
	if len(values) == 0 {
		values = []string{"authorization_code"}
	}
	seen := make(map[string]bool, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		grantType := strings.TrimSpace(value)
		switch grantType {
		case "authorization_code", "refresh_token":
		default:
			return nil, authcontracts.NewOAuthInvalidClientMetadata("unsupported grant_type")
		}
		if !seen[grantType] {
			seen[grantType] = true
			normalized = append(normalized, grantType)
		}
	}
	if !seen["authorization_code"] {
		return nil, authcontracts.NewOAuthInvalidClientMetadata("authorization_code grant is required")
	}
	return normalized, nil
}

func normalizeOAuthResponseTypes(values []string) ([]string, error) {
	if len(values) == 0 {
		values = []string{"code"}
	}
	if len(values) != 1 || strings.TrimSpace(values[0]) != "code" {
		return nil, authcontracts.NewOAuthInvalidClientMetadata("only response_type=code is supported")
	}
	return []string{"code"}, nil
}

func normalizeOAuthScope(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return authcontracts.OAuthScopeMCPChallengeRead, nil
	}
	for _, scope := range strings.Fields(raw) {
		if scope != authcontracts.OAuthScopeMCPChallengeRead {
			return "", authcontracts.NewOAuthInvalidScope("unsupported scope")
		}
	}
	return authcontracts.OAuthScopeMCPChallengeRead, nil
}

func (s *oauthService) normalizeRedirectURIs(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, authcontracts.NewOAuthInvalidClientMetadata("redirect_uris is required")
	}
	normalized := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		redirectURI := strings.TrimSpace(value)
		if redirectURI == "" || !s.redirectURIAllowed(redirectURI) {
			return nil, authcontracts.NewOAuthInvalidClientMetadata("redirect_uri is not allowed")
		}
		if !seen[redirectURI] {
			seen[redirectURI] = true
			normalized = append(normalized, redirectURI)
		}
	}
	return normalized, nil
}

func (s *oauthService) redirectURIAllowed(raw string) bool {
	if isLoopbackRedirectURI(raw) {
		return true
	}
	for _, prefix := range s.config.AllowedRedirectURIPrefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" && strings.HasPrefix(raw, prefix) {
			return true
		}
	}
	return false
}

func isLoopbackRedirectURI(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" || parsed.Host == "" || parsed.Fragment != "" {
		return false
	}
	if parsed.Port() == "" {
		return false
	}
	host := parsed.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func generateOAuthClientID() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return "client_" + base64.RawURLEncoding.EncodeToString(buffer), nil
}
