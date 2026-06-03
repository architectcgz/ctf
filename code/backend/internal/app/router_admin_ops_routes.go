package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type adminOpsRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	ops           *composition.OpsModule
}

func registerAdminOpsRoutes(adminOnly *gin.RouterGroup, deps adminOpsRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	adminOnly.GET("/audit-logs", deps.ops.AuditHandler.ListAuditLogs)
	adminOnly.GET("/dashboard", deps.ops.DashboardHandler.GetDashboard)
	adminOnly.GET("/cheat-detection", deps.ops.RiskHandler.GetCheatDetection)
	adminOnly.POST("/notifications",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionAdminOp,
			ResourceType: "notification_batch",
		}),
		deps.ops.NotificationHandler.PublishAdminNotification,
	)
}
