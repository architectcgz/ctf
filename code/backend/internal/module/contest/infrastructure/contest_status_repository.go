package infrastructure

import (
	"context"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *Repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*contestentity.Contest, int64, error) {
	var contests []*contestentity.Contest
	var total int64

	if len(statuses) == 0 {
		return contests, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&contestentity.Contest{}).
		Where("status IN ?", statuses)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("updated_at ASC, id ASC").Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}
