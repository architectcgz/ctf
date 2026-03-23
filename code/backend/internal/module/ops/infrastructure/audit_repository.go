package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	opsapp "ctf-platform/internal/module/ops/application"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditRepository) List(ctx context.Context, filter opsapp.AuditLogListFilter) ([]opsapp.AuditLogRecord, int64, error) {
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

	records := make([]opsapp.AuditLogRecord, 0)
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
