package container

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
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

func (r *Repository) FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("user_id = ? AND challenge_id = ? AND status IN ?", userID, challengeID,
		[]string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&model.Instance{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) UpdateRuntime(instance *model.Instance) error {
	return r.db.Model(&model.Instance{}).
		Where("id = ?", instance.ID).
		Updates(map[string]any{
			"container_id": instance.ContainerID,
			"network_id":   instance.NetworkID,
			"access_url":   instance.AccessURL,
			"status":       instance.Status,
			"updated_at":   time.Now(),
		}).Error
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

func (r *Repository) AtomicExtend(id int64, userID int64, maxExtends int, duration time.Duration) error {
	result := r.db.Model(&model.Instance{}).
		Where("id = ? AND user_id = ? AND status = ? AND extend_count < ?",
			id, userID, model.InstanceStatusRunning, maxExtends).
		Updates(map[string]interface{}{
			"expires_at":   gorm.Expr("expires_at + ?", duration),
			"extend_count": gorm.Expr("extend_count + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errcode.ErrExtendLimitExceeded
	}
	return nil
}

func (r *Repository) CountRunning() (int64, error) {
	var count int64
	err := r.db.Model(&model.Instance{}).
		Where("status = ?", model.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}

func (r *Repository) ListAllocatedPorts() ([]int, error) {
	var accessURLs []string
	if err := r.db.Model(&model.Instance{}).
		Where("status IN ?", []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Where("access_url <> ''").
		Pluck("access_url", &accessURLs).Error; err != nil {
		return nil, err
	}

	ports := make([]int, 0, len(accessURLs))
	for _, rawURL := range accessURLs {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		portValue := parsed.Port()
		if portValue == "" {
			continue
		}
		port, err := strconv.Atoi(portValue)
		if err != nil {
			continue
		}
		ports = append(ports, port)
	}
	return ports, nil
}
