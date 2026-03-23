package ops

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditlog.Entry) error
}

type AuditLogHandler interface {
	ListAuditLogs(c *gin.Context)
}

type DashboardHandler interface {
	GetDashboard(c *gin.Context)
}

type RiskHandler interface {
	GetCheatDetection(c *gin.Context)
}
