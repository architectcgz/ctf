package ops

import (
	"context"
	"time"

	"gorm.io/gorm"

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

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditRepository) List(ctx context.Context, filter AuditLogListFilter) ([]AuditLogRecord, int64, error) {
	query := r.db.WithContext(ctx).
		Table("audit_logs").
		Joins("LEFT JOIN users ON users.id = audit_logs.user_id")

	if filter.UserID != nil {
		query = query.Where("audit_logs.user_id = ?", *filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("audit_logs.action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Where("audit_logs.resource_type = ?", filter.ResourceType)
	}
	if filter.ResourceID != nil {
		query = query.Where("audit_logs.resource_id = ?", *filter.ResourceID)
	}
	if filter.StartTime != nil {
		query = query.Where("audit_logs.created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("audit_logs.created_at <= ?", *filter.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	records := make([]AuditLogRecord, 0)
	err := query.
		Select(`
			audit_logs.id,
			audit_logs.user_id,
			audit_logs.action,
			audit_logs.resource_type,
			audit_logs.resource_id,
			audit_logs.detail,
			audit_logs.ip_address,
			audit_logs.user_agent,
			audit_logs.created_at,
			COALESCE(users.username, '') AS actor_username
		`).
		Order("audit_logs.created_at DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
