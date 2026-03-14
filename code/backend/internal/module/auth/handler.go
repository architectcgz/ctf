package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service       Service
	tokenService  TokenService
	casProvider   CASProvider
	cookieConfig  CookieConfig
	log           *zap.Logger
	auditRecorder auditlog.Recorder
}

type CookieConfig struct {
	Name     string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
	MaxAge   time.Duration
}

func NewHandler(service Service, tokenService TokenService, casProvider CASProvider, cookieConfig CookieConfig, log *zap.Logger, auditRecorder auditlog.Recorder) *Handler {
	if log == nil {
		log = zap.NewNop()
	}
	if casProvider == nil {
		casProvider = NewCASProvider(config.CASConfig{}, nil, nil, log.Named("cas_provider"), nil)
	}

	return &Handler{
		service:       service,
		tokenService:  tokenService,
		casProvider:   casProvider,
		cookieConfig:  cookieConfig,
		log:           log,
		auditRecorder: auditRecorder,
	}
}

func (h *Handler) Register(c *gin.Context) {
	req := &dto.RegisterReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, tokens, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	h.writeRefreshCookie(c, tokens.RefreshToken)
	response.Success(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	req := &dto.LoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, tokens, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		h.recordAudit(c, auditlog.Entry{
			Action:       model.AuditActionLogin,
			ResourceType: "auth",
			Detail: map[string]any{
				"username":   req.Username,
				"result":     "failed",
				"error":      err.Error(),
				"request_id": c.GetString("request_id"),
			},
			IPAddress: c.ClientIP(),
			UserAgent: userAgentPtr(c.Request.UserAgent()),
		})
		response.FromError(c, err)
		return
	}

	h.writeRefreshCookie(c, tokens.RefreshToken)
	userID := resp.User.ID
	h.recordAudit(c, auditlog.Entry{
		UserID:       &userID,
		Action:       model.AuditActionLogin,
		ResourceType: "auth",
		Detail: map[string]any{
			"username":   resp.User.Username,
			"result":     "success",
			"request_id": c.GetString("request_id"),
		},
		IPAddress: c.ClientIP(),
		UserAgent: userAgentPtr(c.Request.UserAgent()),
	})
	response.Success(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cookieConfig.Name)
	if err != nil {
		response.Error(c, errcode.ErrRefreshTokenExpired)
		return
	}

	payload, refreshErr := h.tokenService.RefreshAccessToken(c.Request.Context(), refreshToken)
	if refreshErr != nil {
		response.FromError(c, refreshErr)
		return
	}

	response.Success(c, &dto.RefreshResp{
		AccessToken: payload.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   payload.ExpiresIn,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	h.log.Info("auth_logout_attempt", zap.Int64("user_id", authUser.UserID), zap.String("username", authUser.Username))
	tokenExpiry := time.Until(authUser.ExpiresAt)
	if tokenExpiry < 0 {
		tokenExpiry = 0
	}
	if err := h.tokenService.RevokeToken(c.Request.Context(), authUser.JTI, tokenExpiry); err != nil {
		h.log.Error("auth_logout_failed_revoke_access_token", zap.Int64("user_id", authUser.UserID), zap.Error(err))
		response.FromError(c, errcode.ErrInternal.WithCause(err))
		return
	}

	if refreshToken, cookieErr := c.Cookie(h.cookieConfig.Name); cookieErr == nil {
		if claims, parseErr := h.tokenService.ParseToken(refreshToken); parseErr == nil {
			refreshExpiry := time.Until(claims.ExpiresAt.Time)
			if refreshExpiry < 0 {
				refreshExpiry = 0
			}
			_ = h.tokenService.RevokeToken(c.Request.Context(), claims.ID, refreshExpiry)
		}
	}

	h.clearRefreshCookie(c)
	h.log.Info("auth_logout_succeeded", zap.Int64("user_id", authUser.UserID), zap.String("username", authUser.Username))
	h.recordAudit(c, auditlog.Entry{
		UserID:       &authUser.UserID,
		Action:       model.AuditActionLogout,
		ResourceType: "auth",
		Detail: map[string]any{
			"username":   authUser.Username,
			"request_id": c.GetString("request_id"),
		},
		IPAddress: c.ClientIP(),
		UserAgent: userAgentPtr(c.Request.UserAgent()),
	})
	response.Success(c, nil)
}

func (h *Handler) Profile(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	profile, err := h.service.GetProfile(c.Request.Context(), authUser.UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, profile)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	req := &dto.ChangePasswordReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), authUser.UserID, req); err != nil {
		h.recordAudit(c, auditlog.Entry{
			UserID:       &authUser.UserID,
			Action:       model.AuditActionUpdate,
			ResourceType: "auth_password",
			Detail: map[string]any{
				"username":   authUser.Username,
				"result":     "failed",
				"error":      err.Error(),
				"request_id": c.GetString("request_id"),
			},
			IPAddress: c.ClientIP(),
			UserAgent: userAgentPtr(c.Request.UserAgent()),
		})
		response.FromError(c, err)
		return
	}

	h.recordAudit(c, auditlog.Entry{
		UserID:       &authUser.UserID,
		Action:       model.AuditActionUpdate,
		ResourceType: "auth_password",
		Detail: map[string]any{
			"username":   authUser.Username,
			"result":     "success",
			"request_id": c.GetString("request_id"),
		},
		IPAddress: c.ClientIP(),
		UserAgent: userAgentPtr(c.Request.UserAgent()),
	})
	response.Success(c, nil)
}

func (h *Handler) IssueWSTicket(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	ticket, err := h.tokenService.IssueWSTicket(c.Request.Context(), authUser)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, &dto.WSTicketResp{
		Ticket:    ticket.Ticket,
		ExpiresAt: ticket.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *Handler) CASStatus(c *gin.Context) {
	response.Success(c, h.casProvider.Status())
}

func (h *Handler) CASLogin(c *gin.Context) {
	resp, err := h.casProvider.BuildLogin(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) CASCallback(c *gin.Context) {
	ticket := c.Query("ticket")
	if ticket == "" {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	resp, tokens, err := h.casProvider.Authenticate(c.Request.Context(), ticket)
	if err != nil {
		h.recordAudit(c, auditlog.Entry{
			Action:       model.AuditActionLogin,
			ResourceType: "auth_cas",
			Detail: map[string]any{
				"provider":   casProviderName,
				"result":     "failed",
				"error":      err.Error(),
				"request_id": c.GetString("request_id"),
			},
			IPAddress: c.ClientIP(),
			UserAgent: userAgentPtr(c.Request.UserAgent()),
		})
		response.FromError(c, err)
		return
	}

	h.writeRefreshCookie(c, tokens.RefreshToken)
	userID := resp.User.ID
	h.recordAudit(c, auditlog.Entry{
		UserID:       &userID,
		Action:       model.AuditActionLogin,
		ResourceType: "auth_cas",
		Detail: map[string]any{
			"provider":   casProviderName,
			"username":   resp.User.Username,
			"result":     "success",
			"request_id": c.GetString("request_id"),
		},
		IPAddress: c.ClientIP(),
		UserAgent: userAgentPtr(c.Request.UserAgent()),
	})
	response.Success(c, resp)
}

func (h *Handler) writeRefreshCookie(c *gin.Context, value string) {
	c.SetSameSite(h.cookieConfig.SameSite)
	c.SetCookie(
		h.cookieConfig.Name,
		value,
		int(h.cookieConfig.MaxAge.Seconds()),
		h.cookieConfig.Path,
		"",
		h.cookieConfig.Secure,
		h.cookieConfig.HTTPOnly,
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(h.cookieConfig.SameSite)
	c.SetCookie(h.cookieConfig.Name, "", -1, h.cookieConfig.Path, "", h.cookieConfig.Secure, h.cookieConfig.HTTPOnly)
}

func (h *Handler) recordAudit(c *gin.Context, entry auditlog.Entry) {
	if h.auditRecorder == nil {
		return
	}
	if err := h.auditRecorder.Record(c.Request.Context(), entry); err != nil {
		h.log.Warn("auth_audit_record_failed", zap.String("action", entry.Action), zap.Error(err))
	}
}

func userAgentPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
