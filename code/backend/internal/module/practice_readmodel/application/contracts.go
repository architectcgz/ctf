package application

import (
	"context"

	"ctf-platform/internal/dto"
)

type QueryRepository interface {
	GetUserProgress(ctx context.Context, userID int64) (totalScore int, totalSolved int, err error)
	GetUserRank(ctx context.Context, userID int64) (int, error)
	GetCategoryStats(ctx context.Context, userID int64) ([]dto.CategoryStat, error)
	GetDifficultyStats(ctx context.Context, userID int64) ([]dto.DifficultyStat, error)
	GetUserTimeline(ctx context.Context, userID int64, limit, offset int) ([]dto.TimelineEvent, error)
}
