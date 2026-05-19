package ports

import (
	"context"
	"time"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/evidence"
)

type ClassItem struct {
	Name         string
	StudentCount int64
}

type StudentItem struct {
	ID               int64
	Username         string
	StudentNo        *string
	Name             *string
	ClassName        *string
	SolvedCount      int
	TotalScore       int
	RecentEventCount int
	WeakDimension    *string
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
	Type         string
	Source       string
	Stage        string
	UserID       int64
	TeamID       *int64
	ChallengeID  int64
	ContestID    *int64
	RoundID      *int64
	ServiceID    *int64
	VictimTeamID *int64
	Title        string
	Timestamp    time.Time
	Detail       string
	Meta         map[string]any
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

type OverviewTrendPoint struct {
	Date               string
	ActiveStudentCount int64
	EventCount         int64
	SolveCount         int64
}

type OverviewTrend struct {
	Points []OverviewTrendPoint
}

type TeachingUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error)
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
	GetStudentEvidence(ctx context.Context, userID int64, query evidence.Query) ([]EvidenceEventRecord, error)
}

type TeachingClassInsightRepository interface {
	GetClassSummary(ctx context.Context, className string, since time.Time) (*ClassSummary, error)
	GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*ClassTrend, error)
	ListClassTeachingFactSnapshots(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error)
}

type TeachingOverviewRepository interface {
	ListStudentsByClasses(ctx context.Context, classNames []string, keyword, studentNo string, since time.Time) ([]StudentItem, error)
	GetOverviewTrend(ctx context.Context, classNames []string, since time.Time, days int) (*OverviewTrend, error)
}
