package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where("id = ? AND status != ?", id, status).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		exists, err := r.contestExists(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return contestdomain.ErrContestNotFound
		}
	}

	return nil
}

func (r *Repository) contestExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Model(&model.Contest{}).
		Select("1").Where("id = ?", id).Limit(1).Find(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}
