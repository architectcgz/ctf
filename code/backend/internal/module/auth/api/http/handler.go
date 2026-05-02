package http

import (
	"context"
	stdhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type authCommandService interface {
	Register(ctx context.Context, req authcmd.RegisterInput) (*dto.LoginResp, *authcontracts.Session, error)
	Login(ctx context.Context, req authcmd.LoginInput) (*dto.LoginResp, *authcontracts.Session, error)
}

type profileCommandService interface {
	ChangePassword(ctx context.Context, userID int64, req identitycontracts.ChangePasswordInput) error
}

type profileQueryService interface {
	GetProfile(ctx context.Context, userID int64) (*dto.AuthUser, error)
}

type casCommandService interface {
	Authenticate(ctx context.Context, ticket string) (*dto.LoginResp, *authcontracts.Session, error)
}

type casQueryService interface {
	Status() *dto.CASStatusResp
	BuildLogin(ctx context.Context) (*dto.CASLoginResp, error)
}

type Handler struct {
	commands      authCommandService
	profileCmd    profileCommandService
	profileQuery  profileQueryService
	casCommands   casCommandService
	casQueries    casQueryService
	tokenService  authcontracts.TokenService
	cookieConfig  CookieConfig
	log           *zap.Logger
	auditRecorder auditlog.Recorder
}

type CookieConfig struct {
	Name     string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite stdhttp.SameSite
	MaxAge   time.Duration
}

const casProviderName = "cas"

func NewHandler(commands authCommandService, profileCmd profileCommandService, profileQuery profileQueryService, tokenService authcontracts.TokenService, casCommands casCommandService, casQueries casQueryService, cookieConfig CookieConfig, log *zap.Logger, auditRecorder auditlog.Recorder) *Handler {
	if log == nil {
		log = zap.NewNop()
	}
	if casCommands == nil {
		casCommands = authcmd.NewCASService(config.CASConfig{}, nil, nil, log.Named("cas_command_service"), nil)
	}
	if casQueries == nil {
		casQueries = authqry.NewCASService(config.CASConfig{})
	}

	return &Handler{
		commands:      commands,
		profileCmd:    profileCmd,
		profileQuery:  profileQuery,
		casCommands:   casCommands,
		casQueries:    casQueries,
		tokenService:  tokenService,
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

	resp, session, err := h.commands.Register(c.Request.Context(), authRequestMapper.ToRegisterInput(*req))
	if err != nil {
		response.FromError(c, err)
		return
	}

	h.writeSessionCookie(c, session.ID)
	response.Success(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	req := &dto.LoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, session, err := h.commands.Login(c.Request.Context(), authRequestMapper.ToLoginInput(*req))
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

	h.writeSessionCookie(c, session.ID)
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

func (h *Handler) Logout(c *gin.Context) {
	authUser := authctx.MustCurrentUser(c)
	h.log.Info("auth_logout_attempt", zap.Int64("user_id", authUser.UserID), zap.String("username", authUser.Username))
	if err := h.tokenService.DeleteSession(c.Request.Context(), authUser.SessionID); err != nil {
		h.log.Error("auth_logout_failed_delete_session", zap.Int64("user_id", authUser.UserID), zap.Error(err))
		response.FromError(c, errcode.ErrInternal.WithCause(err))
		return
	}

	h.clearSessionCookie(c)
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
	if h.profileQuery == nil {
		response.Error(c, errcode.ErrServiceUnavailable)
		return
	}

	authUser := authctx.MustCurrentUser(c)
	profile, err := h.profileQuery.GetProfile(c.Request.Context(), authUser.UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, profile)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	if h.profileCmd == nil {
		response.Error(c, errcode.ErrServiceUnavailable)
		return
	}

	authUser := authctx.MustCurrentUser(c)
	req := &dto.ChangePasswordReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.profileCmd.ChangePassword(c.Request.Context(), authUser.UserID, authRequestMapper.ToChangePasswordInput(*req)); err != nil {
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
	response.Success(c, h.casQueries.Status())
}

func (h *Handler) CASLogin(c *gin.Context) {
	resp, err := h.casQueries.BuildLogin(c.Request.Context())
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

	resp, session, err := h.casCommands.Authenticate(c.Request.Context(), ticket)
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

	h.writeSessionCookie(c, session.ID)
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

func (h *Handler) writeSessionCookie(c *gin.Context, value string) {
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

func (h *Handler) clearSessionCookie(c *gin.Context) {
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
