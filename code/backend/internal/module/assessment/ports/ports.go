package ports

import (
	"context"
	"errors"
	"time"

	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/evidence"
)

var ErrAssessmentReportNotFound = errors.New("assessment report not found")
var ErrAssessmentContestNotFound = errors.New("assessment contest not found")

type AssessmentProfileLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error)
}

type AssessmentProfileReadRepository interface {
	FindByUserID(ctx context.Context, userID int64) ([]*assessmententity.SkillProfile, error)
}

type AssessmentProfileWriteRepository interface {
	Upsert(ctx context.Context, profile *assessmententity.SkillProfile) error
	BatchUpsert(ctx context.Context, profiles []*assessmententity.SkillProfile) error
}

type AssessmentProfileRebuildRepository interface {
	ListStudentIDs(ctx context.Context) ([]int64, error)
}

type AssessmentDimensionScoreRepository interface {
	GetDimensionScores(ctx context.Context, userID int64) ([]assessmentdomain.DimensionScore, error)
	GetDimensionScore(ctx context.Context, userID int64, dimension string) (*assessmentdomain.DimensionScore, error)
}

type AssessmentProfileLockLease interface {
	Release(ctx context.Context) (bool, error)
}

type AssessmentProfileLockStore interface {
	AcquireDimensionUpdateLock(ctx context.Context, userID int64, dimension string, ttl time.Duration) (AssessmentProfileLockLease, bool, error)
	AcquireFullProfileRebuildLock(ctx context.Context, userID int64, ttl time.Duration) (AssessmentProfileLockLease, bool, error)
}

type RecommendationProfileRepository interface {
	FindByUserID(ctx context.Context, userID int64) ([]*assessmententity.SkillProfile, error)
}

type RecommendationTeachingFactRepository interface {
	GetStudentTeachingFactSnapshot(ctx context.Context, userID int64) (*teachingadvice.StudentFactSnapshot, error)
}

type RecommendationSolvedChallengeRepository interface {
	ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error)
}

type RecommendationChallengeRepository interface {
	FindPublishedForRecommendation(ctx context.Context, limit int, dimensions []string, preferredDifficulty string, excludeSolved []int64) ([]*challengecontracts.RecommendationChallenge, error)
}

type AssessmentRecommendationCacheStore interface {
	LoadRecommendations(ctx context.Context, userID int64) ([]*assessmentcontracts.ChallengeRecommendation, bool, error)
	StoreRecommendations(ctx context.Context, userID int64, recommendations []*assessmentcontracts.ChallengeRecommendation, ttl time.Duration) error
	DeleteRecommendations(ctx context.Context, userID int64) error
}

type AssessmentDimensionTotalCacheStore interface {
	LoadPublishedDimensionTotals(ctx context.Context) (map[string]int, bool, error)
	StorePublishedDimensionTotals(ctx context.Context, totals map[string]int, ttl time.Duration) error
	DeletePublishedDimensionTotals(ctx context.Context) error
}

type ProfileRepository interface {
	AssessmentProfileLookupRepository
	AssessmentProfileReadRepository
	AssessmentProfileWriteRepository
	AssessmentProfileRebuildRepository
	AssessmentDimensionScoreRepository
}

type RecommendationRepository interface {
	RecommendationProfileRepository
	RecommendationTeachingFactRepository
	RecommendationSolvedChallengeRepository
}

type ChallengeRepository interface {
	RecommendationChallengeRepository
}

type AssessmentReportLifecycleRepository interface {
	Create(ctx context.Context, report *assessmententity.Report) error
	FindByID(ctx context.Context, reportID int64) (*assessmententity.Report, error)
	MarkReady(ctx context.Context, reportID int64, filePath string, expiresAt time.Time) error
	MarkFailed(ctx context.Context, reportID int64, message string) error
}

type AssessmentReportUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*assessmentdomain.ReportUser, error)
}

type AssessmentReportContestLookupRepository interface {
	FindContestByID(ctx context.Context, contestID int64) (*contestcontracts.Contest, error)
}

type AssessmentPersonalReportRepository interface {
	GetPersonalStats(ctx context.Context, userID int64) (*assessmentdomain.PersonalReportStats, error)
	ListPersonalDimensionStats(ctx context.Context, userID int64) ([]assessmentdomain.ReportDimensionStat, error)
}

type AssessmentClassReportRepository interface {
	CountClassStudents(ctx context.Context, className string) (int, error)
	GetClassAverageScore(ctx context.Context, className string) (float64, error)
	ListClassDimensionAverages(ctx context.Context, className string) ([]assessmentdomain.ClassDimensionAverage, error)
	ListClassTopStudents(ctx context.Context, className string, limit int) ([]assessmentdomain.ClassTopStudent, error)
	ListClassCategoryDistribution(ctx context.Context, className string) ([]assessmentdomain.ClassDistributionStat, error)
	ListClassDifficultyDistribution(ctx context.Context, className string) ([]assessmentdomain.ClassDistributionStat, error)
	GetClassContestMigrationSummary(ctx context.Context, className string) (*assessmentdomain.ClassContestMigrationSummary, error)
}

type ClassInsightSummary struct {
	ClassName          string
	StudentCount       int64
	AverageSolved      float64
	ActiveStudentCount int64
	ActiveRate         float64
	RecentEventCount   int64
}

type ClassInsightTrendPoint struct {
	Date               string
	ActiveStudentCount int64
	EventCount         int64
	SolveCount         int64
}

type ClassInsightTrend struct {
	ClassName string
	Points    []ClassInsightTrendPoint
}

type AssessmentClassInsightRepository interface {
	GetClassSummary(ctx context.Context, className string, since time.Time) (*ClassInsightSummary, error)
	GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*ClassInsightTrend, error)
	ListClassTeachingFactSnapshots(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error)
}

type AssessmentContestExportRepository interface {
	ListContestScoreboard(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportScoreboardItem, error)
	ListContestChallenges(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportChallengeItem, error)
	ListContestTeams(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportTeamItem, error)
}

type AssessmentReviewArchiveRepository interface {
	CountPublishedChallenges(ctx context.Context) (int64, error)
	GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]assessmentdomain.ReviewArchiveTimelineEvent, error)
	GetStudentEvidence(ctx context.Context, userID int64, query evidence.Query) ([]assessmentdomain.ReviewArchiveEvidenceEvent, error)
	ListStudentWriteups(ctx context.Context, userID int64) ([]assessmentdomain.ReviewArchiveWriteupItem, error)
	ListStudentManualReviews(ctx context.Context, userID int64) ([]assessmentdomain.ReviewArchiveManualReviewItem, error)
}

type AssessmentProfileReader interface {
	GetSkillProfile(ctx context.Context, userID int64) (*assessmentcontracts.SkillProfile, error)
}

type ReportRepository interface {
	AssessmentReportLifecycleRepository
	AssessmentReportUserLookupRepository
	AssessmentReportContestLookupRepository
	AssessmentPersonalReportRepository
	AssessmentClassReportRepository
	AssessmentContestExportRepository
	AssessmentReviewArchiveRepository
}
