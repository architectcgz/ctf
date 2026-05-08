package ports

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/evidence"
)

type AssessmentProfileLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type AssessmentProfileReadRepository interface {
	FindByUserID(ctx context.Context, userID int64) ([]*model.SkillProfile, error)
}

type AssessmentProfileWriteRepository interface {
	Upsert(ctx context.Context, profile *model.SkillProfile) error
	BatchUpsert(ctx context.Context, profiles []*model.SkillProfile) error
}

type AssessmentProfileRebuildRepository interface {
	ListStudentIDs(ctx context.Context) ([]int64, error)
}

type AssessmentDimensionScoreRepository interface {
	GetDimensionScores(ctx context.Context, userID int64) ([]assessmentdomain.DimensionScore, error)
	GetDimensionScore(ctx context.Context, userID int64, dimension string) (*assessmentdomain.DimensionScore, error)
}

type RecommendationProfileRepository interface {
	FindByUserID(ctx context.Context, userID int64) ([]*model.SkillProfile, error)
}

type RecommendationTeachingFactRepository interface {
	GetStudentTeachingFactSnapshot(ctx context.Context, userID int64) (*teachingadvice.StudentFactSnapshot, error)
}

type RecommendationSolvedChallengeRepository interface {
	ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error)
}

type RecommendationChallengeRepository interface {
	FindPublishedForRecommendation(ctx context.Context, limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error)
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
	Create(ctx context.Context, report *model.Report) error
	FindByID(ctx context.Context, reportID int64) (*model.Report, error)
	MarkReady(ctx context.Context, reportID int64, filePath string, expiresAt time.Time) error
	MarkFailed(ctx context.Context, reportID int64, message string) error
}

type AssessmentReportUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*assessmentdomain.ReportUser, error)
}

type AssessmentReportContestLookupRepository interface {
	FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error)
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
	GetSkillProfile(ctx context.Context, userID int64) (*dto.SkillProfileResp, error)
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
