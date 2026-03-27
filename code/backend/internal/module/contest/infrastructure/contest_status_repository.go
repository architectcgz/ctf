package infrastructure

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/model"
)

func (r *Repository) ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error) {
	var contests []*model.Contest
	var total int64

	conditions := make([]string, 0, len(statuses))
	args := make([]any, 0, len(statuses)*2)
	for _, status := range statuses {
		switch status {
		case model.ContestStatusRegistration:
			conditions = append(conditions, "(status = ? AND start_time <= ?)")
			args = append(args, status, now)
		case model.ContestStatusRunning:
			conditions = append(conditions, "(status = ? AND ((freeze_time IS NOT NULL AND freeze_time <= ?) OR end_time <= ?))")
			args = append(args, status, now, now)
		case model.ContestStatusFrozen:
			conditions = append(conditions, "(status = ? AND end_time <= ?)")
			args = append(args, status, now)
		}
	}
	if len(conditions) == 0 {
		return contests, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where(strings.Join(conditions, " OR "), args...)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&contests).Error
	return contests, total, err
}
