package ports

import (
	"context"
	"time"
)

type CategoryProgressStat struct {
	Category string `gorm:"column:category"`
	Solved   int    `gorm:"column:solved"`
	Total    int    `gorm:"column:total"`
}

type DifficultyProgressStat struct {
	Difficulty string `gorm:"column:difficulty"`
	Solved     int    `gorm:"column:solved"`
	Total      int    `gorm:"column:total"`
}

type TimelineEventRecord struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   *bool
	Points      *int
	Detail      string
}

type ProgressQueryRepository interface {
	GetUserProgress(ctx context.Context, userID int64) (totalScore int, totalSolved int, err error)
	GetUserRank(ctx context.Context, userID int64) (int, error)
	GetCategoryStats(ctx context.Context, userID int64) ([]CategoryProgressStat, error)
	GetDifficultyStats(ctx context.Context, userID int64) ([]DifficultyProgressStat, error)
}

type TimelineQueryRepository interface {
	GetUserTimeline(ctx context.Context, userID int64, limit, offset int) ([]TimelineEventRecord, error)
}
