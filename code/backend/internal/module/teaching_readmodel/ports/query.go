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
	ClassName        *string `gorm:"column:class_name"`
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

type EvidenceEventRecord struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	Detail      string
	Meta        map[string]any
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

type TeachingUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type TeachingClassQueryRepository interface {
	CountStudentsByClass(ctx context.Context, className string) (int64, error)
	CountClasses(ctx context.Context) (int64, error)
	ListClasses(ctx context.Context, offset, limit int) ([]ClassItem, error)
}

type TeachingStudentDirectoryRepository interface {
	ListStudents(ctx context.Context, className, keyword, studentNo, sortKey, sortOrder string, since time.Time, offset, limit int) ([]StudentItem, int64, error)
	ListStudentsByClass(ctx context.Context, className, keyword, studentNo string, since time.Time) ([]StudentItem, error)
}

type TeachingStudentProfileRepository interface {
	CountPublishedChallenges(ctx context.Context) (int64, error)
	CountSolvedChallenges(ctx context.Context, userID int64) (int64, error)
	GetCategoryProgress(ctx context.Context, userID int64) ([]ProgressRow, error)
	GetDifficultyProgress(ctx context.Context, userID int64) ([]ProgressRow, error)
}

type TeachingStudentActivityRepository interface {
	GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]TimelineEventRecord, error)
	GetStudentEvidence(ctx context.Context, userID int64, challengeID *int64) ([]EvidenceEventRecord, error)
}

type TeachingClassInsightRepository interface {
	GetClassSummary(ctx context.Context, className string, since time.Time) (*ClassSummary, error)
	GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*ClassTrend, error)
}
