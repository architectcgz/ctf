package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) CountRunningInstances(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status = ?", instancecontracts.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}
