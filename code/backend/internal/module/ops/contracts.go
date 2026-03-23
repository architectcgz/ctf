package ops

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditlog.Entry) error
}

type ManagedContainerStat struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   int64
	MemoryLimit   int64
}

type AuditLogService interface {
	AuditRecorder
	ListAuditLogs(ctx context.Context, query *dto.AuditLogQuery) ([]dto.AuditLogItem, int64, int, int, error)
}

type DashboardService interface {
	GetDashboardStats(ctx context.Context) (*dto.DashboardStats, error)
}

type RiskService interface {
	GetCheatDetection(ctx context.Context) (*dto.CheatDetectionResp, error)
}

type NotificationService interface {
	SendNotification(ctx context.Context, userID int64, req *dto.NotificationReq) error
	GetNotifications(ctx context.Context, userID int64, query *dto.NotificationQuery) ([]dto.NotificationInfo, int64, int, int, error)
	MarkAsRead(ctx context.Context, userID, notificationID int64) error
}

type WSTicketConsumer interface {
	ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error)
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

type NotificationHandler interface {
	ListNotifications(c *gin.Context)
	MarkAsRead(c *gin.Context)
	ServeWS(c *gin.Context)
}
