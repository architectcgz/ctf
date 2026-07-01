package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authcmd "ctf-platform/internal/module/auth/application/commands"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

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
