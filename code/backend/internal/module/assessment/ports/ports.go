package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

type ProfileRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	Upsert(ctx context.Context, profile *model.SkillProfile) error
	FindByUserID(ctx context.Context, userID int64) ([]*model.SkillProfile, error)
	ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error)
	BatchUpsert(ctx context.Context, profiles []*model.SkillProfile) error
	ListStudentIDs(ctx context.Context) ([]int64, error)
	GetDimensionScores(ctx context.Context, userID int64) ([]assessmentdomain.DimensionScore, error)
	GetDimensionScore(ctx context.Context, userID int64, dimension string) (*assessmentdomain.DimensionScore, error)
}

type RecommendationRepository interface {
	FindByUserID(ctx context.Context, userID int64) ([]*model.SkillProfile, error)
	ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error)
}

type ChallengeRepository interface {
	FindPublishedForRecommendation(ctx context.Context, limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error)
}

type ReportRepository interface {
	Create(ctx context.Context, report *model.Report) error
	FindByID(ctx context.Context, reportID int64) (*model.Report, error)
	MarkReady(ctx context.Context, reportID int64, filePath string, expiresAt time.Time) error
	MarkFailed(ctx context.Context, reportID int64, message string) error
	FindUserByID(ctx context.Context, userID int64) (*assessmentdomain.ReportUser, error)
	FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error)
	GetPersonalStats(ctx context.Context, userID int64) (*assessmentdomain.PersonalReportStats, error)
	ListPersonalDimensionStats(ctx context.Context, userID int64) ([]assessmentdomain.ReportDimensionStat, error)
	CountClassStudents(ctx context.Context, className string) (int, error)
	GetClassAverageScore(ctx context.Context, className string) (float64, error)
	ListClassDimensionAverages(ctx context.Context, className string) ([]assessmentdomain.ClassDimensionAverage, error)
	ListClassTopStudents(ctx context.Context, className string, limit int) ([]assessmentdomain.ClassTopStudent, error)
	ListContestScoreboard(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportScoreboardItem, error)
	ListContestChallenges(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportChallengeItem, error)
	ListContestTeams(ctx context.Context, contestID int64) ([]assessmentdomain.ContestExportTeamItem, error)
	CountPublishedChallenges(ctx context.Context) (int64, error)
	GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]assessmentdomain.ReviewArchiveTimelineEvent, error)
	GetStudentEvidence(ctx context.Context, userID int64, challengeID *int64) ([]assessmentdomain.ReviewArchiveEvidenceEvent, error)
	ListStudentWriteups(ctx context.Context, userID int64) ([]assessmentdomain.ReviewArchiveWriteupItem, error)
	ListStudentManualReviews(ctx context.Context, userID int64) ([]assessmentdomain.ReviewArchiveManualReviewItem, error)
}

type AssessmentProfileReader interface {
	GetSkillProfile(ctx context.Context, userID int64) (*dto.SkillProfileResp, error)
}
