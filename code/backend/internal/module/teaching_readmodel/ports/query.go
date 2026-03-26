package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type ClassItem struct {
	Name         string `gorm:"column:name"`
	StudentCount int64  `gorm:"column:student_count"`
}

type StudentItem struct {
	ID               int64   `gorm:"column:id"`
	Username         string  `gorm:"column:username"`
	StudentNo        *string `gorm:"column:student_no"`
	Name             *string `gorm:"column:name"`
	SolvedCount      int     `gorm:"column:solved_count"`
	TotalScore       int     `gorm:"column:total_score"`
	RecentEventCount int     `gorm:"column:recent_event_count"`
	WeakDimension    *string `gorm:"column:weak_dimension"`
}

type ProgressRow struct {
	Key    string
	Total  int
	Solved int
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

type ClassSummary struct {
	ClassName          string
	StudentCount       int64
	AverageSolved      float64
	ActiveStudentCount int64
	ActiveRate         float64
	RecentEventCount   int64
}

type ClassTrendPoint struct {
	Date               string
	ActiveStudentCount int64
	EventCount         int64
	SolveCount         int64
}

type ClassTrend struct {
	ClassName string
	Points    []ClassTrendPoint
}

type Repository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	CountStudentsByClass(ctx context.Context, className string) (int64, error)
	ListClasses(ctx context.Context) ([]ClassItem, error)
	ListStudentsByClass(ctx context.Context, className, keyword, studentNo string, since time.Time) ([]StudentItem, error)
	CountPublishedChallenges(ctx context.Context) (int64, error)
	CountSolvedChallenges(ctx context.Context, userID int64) (int64, error)
	GetCategoryProgress(ctx context.Context, userID int64) ([]ProgressRow, error)
	GetDifficultyProgress(ctx context.Context, userID int64) ([]ProgressRow, error)
	GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]TimelineEventRecord, error)
	GetClassSummary(ctx context.Context, className string, since time.Time) (*ClassSummary, error)
	GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*ClassTrend, error)
}
