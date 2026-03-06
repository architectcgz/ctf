package contest

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

var (
	ErrContestNotFound = errors.New("contest not found")
)

type Repository interface {
	Create(ctx context.Context, contest *model.Contest) error
	FindByID(ctx context.Context, id int64) (*model.Contest, error)
	Update(ctx context.Context, contest *model.Contest) error
	List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error)
	ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, contest *model.Contest) error {
	return r.db.WithContext(ctx).Create(contest).Error
}

func (r *repository) FindByID(ctx context.Context, id int64) (*model.Contest, error) {
	var contest model.Contest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&contest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContestNotFound
		}
		return nil, err
	}
	return &contest, nil
}

func (r *repository) Update(ctx context.Context, contest *model.Contest) error {
	return r.db.WithContext(ctx).Save(contest).Error
}

func (r *repository) List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error) {
	var contests []*model.Contest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Contest{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}

func (r *repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error) {
	var contests []*model.Contest
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Contest{}).Where("status IN ?", statuses).
		Where(r.db.Where("status = ? AND start_time <= ?", model.ContestStatusRegistration, now).
			Or("status = ? AND end_time <= ?", model.ContestStatusRunning, now))

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where("id = ? AND status != ?", id, status).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	// RowsAffected == 0 可能是不存在或状态已相同，需要区分
	if result.RowsAffected == 0 {
		var exists bool
		err := r.db.WithContext(ctx).Model(&model.Contest{}).
			Select("1").Where("id = ?", id).Limit(1).Find(&exists).Error
		if err != nil {
			return err
		}
		if !exists {
			return ErrContestNotFound
		}
	}

	return nil
}
