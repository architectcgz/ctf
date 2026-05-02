package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
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

type AuditCommandRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
}

type AuditQueryRepository interface {
	List(ctx context.Context, filter AuditLogListFilter) ([]AuditLogRecord, int64, error)
}
