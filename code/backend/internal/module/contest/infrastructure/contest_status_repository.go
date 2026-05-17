package infrastructure

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

func (r *Repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error) {
	var contests []*model.Contest
	var total int64

	if len(statuses) == 0 {
		return contests, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where("status IN ?", statuses)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("updated_at ASC, id ASC").Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}
