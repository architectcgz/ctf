package http

import (
	"bytes"
	"errors"
	"html/template"
	stdhttp "net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

var oauthConsentPageTemplate = template.Must(template.New("oauth_consent").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CTF MCP 授权</title>
</head>
<body>
  <main>
    <h1>授权 {{ .ClientName }}</h1>
    <p>当前用户：{{ .Username }}</p>
    <p>请求权限：{{ .Scope }}</p>
    <form method="post" action="/api/v1/oauth/authorize">
      <input type="hidden" name="response_type" value="{{ .ResponseType }}">
      <input type="hidden" name="client_id" value="{{ .ClientID }}">
      <input type="hidden" name="redirect_uri" value="{{ .RedirectURI }}">
      <input type="hidden" name="scope" value="{{ .Scope }}">
      <input type="hidden" name="state" value="{{ .State }}">
      <input type="hidden" name="code_challenge" value="{{ .CodeChallenge }}">
      <input type="hidden" name="code_challenge_method" value="{{ .CodeChallengeMethod }}">
      <input type="hidden" name="csrf_nonce" value="{{ .CSRFNonce }}">
      <button type="submit" name="approve" value="true">允许</button>
      <button type="submit" name="approve" value="false">拒绝</button>
    </form>
  </main>
</body>
</html>`))

type oauthConsentPageData struct {
	ClientName          string
	Username            string
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	CSRFNonce           string
}

func (h *Handler) OAuthProtectedResourceMetadata(c *gin.Context) {
	if h.oauthMetadata == nil {
		c.JSON(stdhttp.StatusServiceUnavailable, OAuthErrorResp{Error: "server_error", ErrorDescription: "oauth metadata service is unavailable"})
		return
	}
	resp, err := h.oauthMetadata.ProtectedResource(c.Request.Context(), requestOrigin(c))
	if err != nil {
		c.JSON(stdhttp.StatusInternalServerError, OAuthErrorResp{Error: "server_error", ErrorDescription: err.Error()})
		return
	}
	c.JSON(stdhttp.StatusOK, resp)
}

func (h *Handler) OAuthAuthorizationServerMetadata(c *gin.Context) {
	if h.oauthMetadata == nil {
		c.JSON(stdhttp.StatusServiceUnavailable, OAuthErrorResp{Error: "server_error", ErrorDescription: "oauth metadata service is unavailable"})
		return
	}
	resp, err := h.oauthMetadata.AuthorizationServer(c.Request.Context(), requestOrigin(c))
	if err != nil {
		c.JSON(stdhttp.StatusInternalServerError, OAuthErrorResp{Error: "server_error", ErrorDescription: err.Error()})
		return
	}
	c.JSON(stdhttp.StatusOK, resp)
}

func (h *Handler) OAuthAuthorize(c *gin.Context) {
	if h.oauthCommands == nil {
		c.JSON(stdhttp.StatusServiceUnavailable, OAuthErrorResp{Error: "server_error", ErrorDescription: "oauth authorization service is unavailable"})
		return
	}
	req := oauthAuthorizationRequestFromQuery(c)
	if _, err := h.oauthCommands.ValidateAuthorizationRequest(c.Request.Context(), req); err != nil {
		writeOAuthError(c, err)
		return
	}

	user, authenticated, err := h.currentOAuthUser(c)
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	if !authenticated {
		c.Redirect(stdhttp.StatusFound, "/login?redirect="+url.QueryEscape(c.Request.URL.RequestURI()))
		return
	}

	result, err := h.oauthCommands.PrepareAuthorization(c.Request.Context(), authcmd.OAuthAuthorizationInput{
		Request: req,
		User:    user,
	})
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	if result.NeedsConsent {
		h.renderOAuthConsent(c, req, user, result)
		return
	}
	c.Redirect(stdhttp.StatusFound, result.RedirectTo)
}

func (h *Handler) OAuthAuthorizeDecision(c *gin.Context) {
	if h.oauthCommands == nil {
		c.JSON(stdhttp.StatusServiceUnavailable, OAuthErrorResp{Error: "server_error", ErrorDescription: "oauth authorization service is unavailable"})
		return
	}
	user, authenticated, err := h.currentOAuthUser(c)
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	if !authenticated {
		c.Redirect(stdhttp.StatusFound, "/login?redirect="+url.QueryEscape(c.Request.URL.RequestURI()))
		return
	}
	input := authcmd.OAuthAuthorizationDecisionInput{
		Request:   oauthAuthorizationRequestFromForm(c),
		User:      user,
		CSRFNonce: c.PostForm("csrf_nonce"),
	}
	var result *authcmd.OAuthAuthorizationResult
	approved := isApprovedConsent(c.PostForm("approve"))
	if approved {
		result, err = h.oauthCommands.ApproveAuthorization(c.Request.Context(), input)
	} else {
		result, err = h.oauthCommands.DenyAuthorization(c.Request.Context(), input)
	}
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	if approved {
		userID := user.UserID
		h.recordAudit(c, auditlog.Entry{
			UserID:       &userID,
			Action:       auditlog.ActionCreate,
			ResourceType: "oauth_consent",
			Detail: map[string]any{
				"event":      "oauth_consent/grant",
				"client_id":  result.Client.ClientID,
				"scope":      result.Scope,
				"result":     "success",
				"request_id": c.GetString("request_id"),
			},
			IPAddress: c.ClientIP(),
			UserAgent: normalizeOptionalString(c.Request.UserAgent()),
		})
	}
	c.Redirect(stdhttp.StatusFound, result.RedirectTo)
}

func (h *Handler) OAuthToken(c *gin.Context) {
	if h.oauthCommands == nil {
		c.JSON(stdhttp.StatusServiceUnavailable, OAuthErrorResp{Error: "server_error", ErrorDescription: "oauth token service is unavailable"})
		return
	}

	grantType := strings.TrimSpace(c.PostForm("grant_type"))
	clientID := strings.TrimSpace(c.PostForm("client_id"))
	var (
		result *authcmd.OAuthTokenResult
		err    error
		event  string
		action string
	)
	switch grantType {
	case "authorization_code":
		event = "oauth_token/exchange"
		action = auditlog.ActionCreate
		result, err = h.oauthCommands.ExchangeAuthorizationCode(c.Request.Context(), authcmd.OAuthAuthorizationCodeExchangeInput{
			ClientID:     clientID,
			Code:         c.PostForm("code"),
			RedirectURI:  c.PostForm("redirect_uri"),
			CodeVerifier: c.PostForm("code_verifier"),
		})
	case "refresh_token":
		event = "oauth_token/refresh"
		action = auditlog.ActionUpdate
		result, err = h.oauthCommands.RefreshAccessToken(c.Request.Context(), authcmd.OAuthRefreshTokenInput{
			ClientID:       clientID,
			RefreshToken:   c.PostForm("refresh_token"),
			RequestedScope: c.PostForm("scope"),
		})
	default:
		event = "oauth_token/unsupported_grant"
		action = auditlog.ActionCreate
		err = authcontracts.NewOAuthInvalidGrant("unsupported grant_type")
	}
	if err != nil {
		h.recordOAuthTokenAudit(c, action, event, clientID, nil, "", err)
		writeOAuthError(c, err)
		return
	}
	userID := result.UserID
	h.recordOAuthTokenAudit(c, action, event, result.ClientID, &userID, result.Scope, nil)
	c.JSON(stdhttp.StatusOK, OAuthTokenResp{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		Scope:        result.Scope,
	})
}

func (h *Handler) RegisterOAuthClient(c *gin.Context) {
	if h.oauthCommands == nil {
		c.JSON(stdhttp.StatusServiceUnavailable, OAuthErrorResp{Error: "server_error", ErrorDescription: "oauth registration service is unavailable"})
		return
	}
	req := &OAuthClientRegistrationReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		writeOAuthError(c, authcontracts.NewOAuthError("invalid_client_metadata", err.Error(), stdhttp.StatusBadRequest))
		return
	}

	resp, err := h.oauthCommands.RegisterClient(c.Request.Context(), authcmd.OAuthClientRegistrationInput{
		ClientName:              req.ClientName,
		ClientURI:               req.ClientURI,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              req.GrantTypes,
		ResponseTypes:           req.ResponseTypes,
		Scope:                   req.Scope,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
	})
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	h.recordAudit(c, auditlog.Entry{
		Action:       auditlog.ActionCreate,
		ResourceType: "oauth_client",
		Detail: map[string]any{
			"event":       "oauth_client/register",
			"client_id":   resp.ClientID,
			"client_name": resp.ClientName,
			"scope":       resp.Scope,
			"result":      "success",
			"request_id":  c.GetString("request_id"),
		},
		IPAddress: c.ClientIP(),
		UserAgent: normalizeOptionalString(c.Request.UserAgent()),
	})

	c.JSON(stdhttp.StatusCreated, OAuthClientRegistrationResp{
		ClientID:                resp.ClientID,
		ClientName:              resp.ClientName,
		ClientURI:               resp.ClientURI,
		RedirectURIs:            resp.RedirectURIs,
		GrantTypes:              resp.GrantTypes,
		ResponseTypes:           resp.ResponseTypes,
		Scope:                   resp.Scope,
		TokenEndpointAuthMethod: resp.TokenEndpointAuthMethod,
	})
}

func (h *Handler) renderOAuthConsent(c *gin.Context, req authcmd.OAuthAuthorizationRequest, user authctx.CurrentUser, result *authcmd.OAuthAuthorizationResult) {
	var page bytes.Buffer
	if err := oauthConsentPageTemplate.Execute(&page, oauthConsentPageData{
		ClientName:          result.Client.ClientName,
		Username:            user.Username,
		ResponseType:        req.ResponseType,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		Scope:               result.Scope,
		State:               req.State,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		CSRFNonce:           result.CSRFNonce,
	}); err != nil {
		c.JSON(stdhttp.StatusInternalServerError, OAuthErrorResp{Error: "server_error", ErrorDescription: err.Error()})
		return
	}
	c.Data(stdhttp.StatusOK, "text/html; charset=utf-8", page.Bytes())
}

func (h *Handler) recordOAuthTokenAudit(c *gin.Context, action, event, clientID string, userID *int64, scope string, cause error) {
	detail := map[string]any{
		"event":      event,
		"client_id":  clientID,
		"request_id": c.GetString("request_id"),
	}
	if scope != "" {
		detail["scope"] = scope
	}
	if cause != nil {
		detail["result"] = "failed"
		detail["error"] = oauthAuditErrorCode(cause)
	} else {
		detail["result"] = "success"
	}
	h.recordAudit(c, auditlog.Entry{
		UserID:       userID,
		Action:       action,
		ResourceType: "oauth_token",
		Detail:       detail,
		IPAddress:    c.ClientIP(),
		UserAgent:    normalizeOptionalString(c.Request.UserAgent()),
	})
}

func oauthAuditErrorCode(err error) string {
	var oauthErr *authcontracts.OAuthError
	if errors.As(err, &oauthErr) {
		return oauthErr.Code
	}
	return "server_error"
}

func (h *Handler) currentOAuthUser(c *gin.Context) (authctx.CurrentUser, bool, error) {
	if h.tokenService == nil {
		return authctx.CurrentUser{}, false, authcontracts.NewOAuthError("server_error", "token service is unavailable", stdhttp.StatusServiceUnavailable)
	}
	sessionID, err := c.Cookie(h.cookieConfig.Name)
	if err != nil || sessionID == "" {
		return authctx.CurrentUser{}, false, nil
	}
	session, err := h.tokenService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		return authctx.CurrentUser{}, false, nil
	}
	return authctx.CurrentUser{
		UserID:    session.UserID,
		Username:  session.Username,
		Role:      session.Role,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt,
	}, true, nil
}

func oauthAuthorizationRequestFromQuery(c *gin.Context) authcmd.OAuthAuthorizationRequest {
	return authcmd.OAuthAuthorizationRequest{
		ResponseType:        c.Query("response_type"),
		ClientID:            c.Query("client_id"),
		RedirectURI:         c.Query("redirect_uri"),
		Scope:               c.Query("scope"),
		State:               c.Query("state"),
		CodeChallenge:       c.Query("code_challenge"),
		CodeChallengeMethod: c.Query("code_challenge_method"),
	}
}

func oauthAuthorizationRequestFromForm(c *gin.Context) authcmd.OAuthAuthorizationRequest {
	return authcmd.OAuthAuthorizationRequest{
		ResponseType:        c.PostForm("response_type"),
		ClientID:            c.PostForm("client_id"),
		RedirectURI:         c.PostForm("redirect_uri"),
		Scope:               c.PostForm("scope"),
		State:               c.PostForm("state"),
		CodeChallenge:       c.PostForm("code_challenge"),
		CodeChallengeMethod: c.PostForm("code_challenge_method"),
	}
}

func isApprovedConsent(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func writeOAuthError(c *gin.Context, err error) {
	var oauthErr *authcontracts.OAuthError
	if errors.As(err, &oauthErr) {
		c.JSON(oauthErr.StatusCode, OAuthErrorResp{
			Error:            oauthErr.Code,
			ErrorDescription: oauthErr.Description,
		})
		return
	}
	c.JSON(stdhttp.StatusInternalServerError, OAuthErrorResp{
		Error:            "server_error",
		ErrorDescription: err.Error(),
	})
}

func requestOrigin(c *gin.Context) string {
	scheme := firstForwardedValue(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = c.Request.URL.Scheme
	}
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := firstForwardedValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = c.Request.Host
	}
	return strings.TrimRight(scheme+"://"+host, "/")
}

func firstForwardedValue(raw string) string {
	if raw == "" {
		return ""
	}
	if idx := strings.Index(raw, ","); idx >= 0 {
		raw = raw[:idx]
	}
	return strings.TrimSpace(raw)
}
