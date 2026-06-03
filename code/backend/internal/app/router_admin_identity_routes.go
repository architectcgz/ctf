package app

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	response "ctf-platform/internal/httpresponse"
	"ctf-platform/internal/middleware"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
)

type adminIdentityRouteDeps struct {
	auditRecorder   auditlog.Recorder
	auditLogger     *zap.Logger
	identityHandler *identityhttp.Handler
	tokenService    authcontracts.TokenService
}

func registerAdminIdentityRoutes(adminOnly *gin.RouterGroup, deps adminIdentityRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	users := adminOnly.Group("/users")
	users.GET("", deps.identityHandler.ListUsers)
	users.POST("",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "user",
		}),
		deps.identityHandler.CreateUser,
	)
	users.POST("/import",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "user_import",
		}),
		deps.identityHandler.ImportUsers,
	)

	userByID := users.Group("/:id")
	userByID.Use(middleware.ParseInt64Param("id"))
	userByID.PUT("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.identityHandler.UpdateUser,
	)
	userByID.DELETE("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.identityHandler.DeleteUser,
	)

	userSessions := userByID.Group("/sessions")
	userSessions.GET("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionRead,
			ResourceType:    "user_session",
			ResourceIDParam: "id",
		}),
		func(c *gin.Context) {
			userID := c.GetInt64("id")
			sessions, err := deps.tokenService.ListUserSessions(c.Request.Context(), userID)
			if err != nil {
				response.FromError(c, err)
				return
			}
			resps := make([]identityhttp.UserSessionResp, 0, len(sessions))
			for _, s := range sessions {
				resps = append(resps, identityhttp.UserSessionResp{
					ID:        s.ID,
					Username:  s.Username,
					Role:      s.Role,
					ExpiresAt: s.ExpiresAt,
				})
			}
			response.Success(c, gin.H{"sessions": resps})
		},
	)
	userSessions.DELETE("/:sid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "user_session",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		func(c *gin.Context) {
			sessionID := c.Param("sid")
			if sessionID == "" {
				response.InvalidParams(c, "缺少会话ID")
				return
			}

			session, err := deps.tokenService.GetSession(c.Request.Context(), sessionID)
			if err != nil {
				if errors.Is(err, authcontracts.ErrAccessTokenExpired) {
					response.Error(c, apperror.ErrNotFound.WithMessage("该会话已不活跃或已被撤销"))
					return
				}
				response.FromError(c, err)
				return
			}
			userID := c.GetInt64("id")
			if session.UserID != userID {
				response.Error(c, apperror.ErrForbidden)
				return
			}
			if err := deps.tokenService.DeleteSession(c.Request.Context(), sessionID); err != nil {
				response.FromError(c, err)
				return
			}
			response.Success(c, gin.H{"message": "会话已撤销"})
		},
	)
	userSessions.DELETE("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "user_session",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		func(c *gin.Context) {
			userID := c.GetInt64("id")
			if err := deps.tokenService.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
				response.FromError(c, err)
				return
			}
			response.Success(c, gin.H{"message": "已撤销该用户所有会话"})
		},
	)
}
