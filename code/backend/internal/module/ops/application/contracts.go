package application

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	opsmodule "ctf-platform/internal/module/ops"
	ctfws "ctf-platform/pkg/websocket"
)

type AuditLogListFilter struct {
	UserID       *int64
	Action       string
	ResourceType string
	ResourceID   *int64
	StartTime    *time.Time
	EndTime      *time.Time
	Offset       int
	Limit        int
}

type AuditLogRecord struct {
	ID            int64     `gorm:"column:id"`
	UserID        *int64    `gorm:"column:user_id"`
	Action        string    `gorm:"column:action"`
	ResourceType  string    `gorm:"column:resource_type"`
	ResourceID    *int64    `gorm:"column:resource_id"`
	Detail        string    `gorm:"column:detail"`
	IPAddress     string    `gorm:"column:ip_address"`
	UserAgent     *string   `gorm:"column:user_agent"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	ActorUsername string    `gorm:"column:actor_username"`
}

type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, filter AuditLogListFilter) ([]AuditLogRecord, int64, error)
}

type RuntimeQuery interface {
	CountRunning() (int64, error)
}

type RuntimeStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]opsmodule.ManagedContainerStat, error)
}

type NotificationListFilter struct {
	UserID int64
	Type   string
	Offset int
	Limit  int
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	List(ctx context.Context, filter NotificationListFilter) ([]model.Notification, int64, error)
	FindByID(ctx context.Context, notificationID, userID int64) (*model.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error
}

type NotificationBroadcaster interface {
	SendToUser(userID int64, message ctfws.Envelope) int
	Broadcast(message ctfws.Envelope) int
}

type RiskAuditEvent struct {
	UserID    *int64
	Username  string
	IPAddress string
	CreatedAt time.Time
}

type RiskRepository interface {
	ListRecentSubmitEvents(ctx context.Context, since time.Time, limit int) ([]RiskAuditEvent, error)
	ListRecentLoginEvents(ctx context.Context, since time.Time, limit int) ([]RiskAuditEvent, error)
}
