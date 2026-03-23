package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	opsapp "ctf-platform/internal/module/ops/application"
)

type RiskRepository struct {
	db *gorm.DB
}

func NewRiskRepository(db *gorm.DB) *RiskRepository {
	return &RiskRepository{db: db}
}

func (r *RiskRepository) ListRecentSubmitEvents(ctx context.Context, since time.Time, limit int) ([]opsapp.RiskAuditEvent, error) {
	events := make([]opsapp.RiskAuditEvent, 0)
	err := r.db.WithContext(ctx).
		Table("audit_logs").
		Select("audit_logs.user_id, users.username, audit_logs.ip_address, audit_logs.created_at").
		Joins("LEFT JOIN users ON users.id = audit_logs.user_id").
		Where("audit_logs.action = ? AND audit_logs.created_at >= ?", model.AuditActionSubmit, since).
		Order("audit_logs.created_at DESC").
		Limit(limit).
		Scan(&events).Error
	return events, err
}

func (r *RiskRepository) ListRecentLoginEvents(ctx context.Context, since time.Time, limit int) ([]opsapp.RiskAuditEvent, error) {
	events := make([]opsapp.RiskAuditEvent, 0)
	err := r.db.WithContext(ctx).
		Table("audit_logs").
		Select("audit_logs.user_id, users.username, audit_logs.ip_address, audit_logs.created_at").
		Joins("LEFT JOIN users ON users.id = audit_logs.user_id").
		Where("audit_logs.action = ? AND audit_logs.created_at >= ?", model.AuditActionLogin, since).
		Order("audit_logs.created_at DESC").
		Limit(limit).
		Scan(&events).Error
	return events, err
}
