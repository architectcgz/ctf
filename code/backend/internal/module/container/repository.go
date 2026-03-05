package container

import (
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(instance *model.Instance) error {
	return r.db.Create(instance).Error
}

func (r *Repository) FindByID(id int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("id = ?", id).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindByUserID(userID int64) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.db.Where("user_id = ? AND status IN ?", userID,
		[]string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&model.Instance{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) FindExpired() ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.db.Where("status = ? AND expires_at < ?",
		model.InstanceStatusRunning, time.Now()).
		Find(&instances).Error
	return instances, err
}

func (r *Repository) UpdateExtend(id int64, expiresAt time.Time, extendCount int) error {
	return r.db.Model(&model.Instance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"expires_at":   expiresAt,
			"extend_count": extendCount,
		}).Error
}
