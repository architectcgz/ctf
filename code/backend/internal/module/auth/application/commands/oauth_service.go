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
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

type OAuthService interface {
	RegisterClient(ctx context.Context, req OAuthClientRegistrationInput) (*OAuthClientRegistrationResp, error)
	ValidateAuthorizationRequest(ctx context.Context, req OAuthAuthorizationRequest) (*OAuthAuthorizationValidation, error)
	PrepareAuthorization(ctx context.Context, input OAuthAuthorizationInput) (*OAuthAuthorizationResult, error)
	ApproveAuthorization(ctx context.Context, input OAuthAuthorizationDecisionInput) (*OAuthAuthorizationResult, error)
	DenyAuthorization(ctx context.Context, input OAuthAuthorizationDecisionInput) (*OAuthAuthorizationResult, error)
}

type OAuthClientStore interface {
	SaveClient(ctx context.Context, client authcontracts.OAuthClient) error
	FindClientByID(ctx context.Context, clientID string) (*authcontracts.OAuthClient, error)
	FindActiveConsent(ctx context.Context, userID int64, clientID, scope string) (*authcontracts.OAuthConsent, error)
	UpsertConsent(ctx context.Context, consent authcontracts.OAuthConsent) error
	StoreAuthorizationCode(ctx context.Context, code string, claims authcontracts.OAuthAuthorizationCode, ttl time.Duration) error
	StoreConsentNonce(ctx context.Context, nonce string, ttl time.Duration) error
	ConsumeConsentNonce(ctx context.Context, nonce string) (bool, error)
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

type OAuthAuthorizationRequest struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
}

type OAuthAuthorizationInput struct {
	Request OAuthAuthorizationRequest
	User    authctx.CurrentUser
}

type OAuthAuthorizationDecisionInput struct {
	Request   OAuthAuthorizationRequest
	User      authctx.CurrentUser
	CSRFNonce string
}

type OAuthAuthorizationValidation struct {
	Client authcontracts.OAuthClient
	Scope  string
}

type OAuthAuthorizationResult struct {
	Client       authcontracts.OAuthClient
	Scope        string
	RedirectURI  string
	State        string
	Code         string
	RedirectTo   string
	CSRFNonce    string
	NeedsConsent bool
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

func (s *oauthService) ValidateAuthorizationRequest(ctx context.Context, req OAuthAuthorizationRequest) (*OAuthAuthorizationValidation, error) {
	if s.store == nil {
		return nil, apperror.ErrServiceUnavailable
	}
	return s.validateAuthorizationRequest(ctx, req)
}

func (s *oauthService) PrepareAuthorization(ctx context.Context, input OAuthAuthorizationInput) (*OAuthAuthorizationResult, error) {
	validation, err := s.validateAuthorizationRequest(ctx, input.Request)
	if err != nil {
		return nil, err
	}
	if input.User.UserID <= 0 {
		return nil, authcontracts.NewOAuthInvalidRequest("user session is required")
	}
	consent, err := s.store.FindActiveConsent(ctx, input.User.UserID, validation.Client.ClientID, validation.Scope)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if consent != nil {
		return s.issueAuthorizationCode(ctx, input, *validation)
	}

	nonce, err := generateOAuthOpaque("nonce_", 24)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if err := s.store.StoreConsentNonce(ctx, nonce, s.config.AuthorizationCodeTTL); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return &OAuthAuthorizationResult{
		Client:       validation.Client,
		Scope:        validation.Scope,
		RedirectURI:  input.Request.RedirectURI,
		State:        input.Request.State,
		CSRFNonce:    nonce,
		NeedsConsent: true,
	}, nil
}

func (s *oauthService) ApproveAuthorization(ctx context.Context, input OAuthAuthorizationDecisionInput) (*OAuthAuthorizationResult, error) {
	if err := s.consumeConsentNonce(ctx, input.CSRFNonce); err != nil {
		return nil, err
	}
	validation, err := s.validateAuthorizationRequest(ctx, input.Request)
	if err != nil {
		return nil, err
	}
	if input.User.UserID <= 0 {
		return nil, authcontracts.NewOAuthInvalidRequest("user session is required")
	}
	if err := s.store.UpsertConsent(ctx, authcontracts.OAuthConsent{
		UserID:    input.User.UserID,
		ClientID:  validation.Client.ClientID,
		Scope:     validation.Scope,
		GrantedAt: time.Now().UTC(),
	}); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return s.issueAuthorizationCode(ctx, OAuthAuthorizationInput{Request: input.Request, User: input.User}, *validation)
}

func (s *oauthService) DenyAuthorization(ctx context.Context, input OAuthAuthorizationDecisionInput) (*OAuthAuthorizationResult, error) {
	if err := s.consumeConsentNonce(ctx, input.CSRFNonce); err != nil {
		return nil, err
	}
	validation, err := s.validateAuthorizationRequest(ctx, input.Request)
	if err != nil {
		return nil, err
	}
	return &OAuthAuthorizationResult{
		Client:      validation.Client,
		Scope:       validation.Scope,
		RedirectURI: input.Request.RedirectURI,
		State:       input.Request.State,
		RedirectTo:  appendOAuthQuery(input.Request.RedirectURI, map[string]string{"error": "access_denied", "state": input.Request.State}),
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

func (s *oauthService) validateAuthorizationRequest(ctx context.Context, req OAuthAuthorizationRequest) (*OAuthAuthorizationValidation, error) {
	if strings.TrimSpace(req.ResponseType) != "code" {
		return nil, authcontracts.NewOAuthInvalidRequest("response_type must be code")
	}
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		return nil, authcontracts.NewOAuthInvalidRequest("client_id is required")
	}
	client, err := s.store.FindClientByID(ctx, clientID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if client == nil {
		return nil, authcontracts.NewOAuthInvalidClient("client is not registered")
	}
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" || !client.AllowsRedirectURI(redirectURI) {
		return nil, authcontracts.NewOAuthInvalidRequest("redirect_uri is not registered")
	}
	if strings.TrimSpace(req.CodeChallenge) == "" {
		return nil, authcontracts.NewOAuthInvalidRequest("code_challenge is required")
	}
	if strings.TrimSpace(req.CodeChallengeMethod) != "S256" {
		return nil, authcontracts.NewOAuthInvalidRequest("code_challenge_method must be S256")
	}
	scope, err := normalizeOAuthScope(req.Scope)
	if err != nil {
		return nil, err
	}
	if !scopeAllowedByClient(scope, client.Scope) {
		return nil, authcontracts.NewOAuthInvalidScope("scope is not registered for client")
	}
	return &OAuthAuthorizationValidation{
		Client: *client,
		Scope:  scope,
	}, nil
}

func (s *oauthService) consumeConsentNonce(ctx context.Context, nonce string) error {
	if strings.TrimSpace(nonce) == "" {
		return authcontracts.NewOAuthInvalidRequest("csrf_nonce is required")
	}
	ok, err := s.store.ConsumeConsentNonce(ctx, nonce)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if !ok {
		return authcontracts.NewOAuthInvalidRequest("csrf_nonce is invalid or expired")
	}
	return nil
}

func (s *oauthService) issueAuthorizationCode(ctx context.Context, input OAuthAuthorizationInput, validation OAuthAuthorizationValidation) (*OAuthAuthorizationResult, error) {
	code, err := generateOAuthOpaque("code_", 32)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	now := time.Now().UTC()
	claims := authcontracts.OAuthAuthorizationCode{
		UserID:              input.User.UserID,
		Username:            input.User.Username,
		Role:                input.User.Role,
		ClientID:            validation.Client.ClientID,
		RedirectURI:         input.Request.RedirectURI,
		Scope:               validation.Scope,
		CodeChallenge:       input.Request.CodeChallenge,
		CodeChallengeMethod: input.Request.CodeChallengeMethod,
		IssuedAt:            now,
		ExpiresAt:           now.Add(s.config.AuthorizationCodeTTL).UTC(),
	}
	if err := s.store.StoreAuthorizationCode(ctx, code, claims, s.config.AuthorizationCodeTTL); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return &OAuthAuthorizationResult{
		Client:      validation.Client,
		Scope:       validation.Scope,
		RedirectURI: input.Request.RedirectURI,
		State:       input.Request.State,
		Code:        code,
		RedirectTo:  appendOAuthQuery(input.Request.RedirectURI, map[string]string{"code": code, "state": input.Request.State}),
	}, nil
}

func scopeAllowedByClient(requestedScope, clientScope string) bool {
	for _, scope := range strings.Fields(clientScope) {
		if scope == requestedScope {
			return true
		}
	}
	return false
}

func appendOAuthQuery(raw string, values map[string]string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := parsed.Query()
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
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
	return generateOAuthOpaque("client_", 24)
}

func generateOAuthOpaque(prefix string, size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(buffer), nil
}
