package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	runtimestate "ctf-platform/internal/module/runtime/contracts"
)

type ManagedInstanceRepository struct {
	db *gorm.DB
}

func NewManagedInstanceRepository(db *gorm.DB) *ManagedInstanceRepository {
	return &ManagedInstanceRepository{db: db}
}

func (r *ManagedInstanceRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ManagedInstanceRepository) FindByID(ctx context.Context, id int64) (*runtimestate.RuntimeManagedInstance, error) {
	var instance runtimestate.RuntimeManagedInstance
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}
