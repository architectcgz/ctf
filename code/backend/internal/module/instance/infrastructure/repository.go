package infrastructure

import (
	"context"

	"gorm.io/gorm"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CountRunningInstances(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status = ?", instancecontracts.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}
