package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (r *Repository) Create(ctx context.Context, contest *model.Contest) error {
	return r.db.WithContext(ctx).Create(contest).Error
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*model.Contest, error) {
	var contest model.Contest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&contest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contestdomain.ErrContestNotFound
		}
		return nil, err
	}
	return &contest, nil
}

func (r *Repository) Update(ctx context.Context, contest *model.Contest) error {
	return r.db.WithContext(ctx).Save(contest).Error
}

func (r *Repository) List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error) {
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
