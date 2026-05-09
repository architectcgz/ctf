package ports

import (
	"context"

	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

type TeacherAWDReviewContestFilter struct {
	Status  string
	Keyword string
	Offset  int
	Limit   int
}

type TeacherAWDReviewContestSummary struct {
	RunningCount     int64
	ExportReadyCount int64
}

type TeacherAWDReviewRepository interface {
	ListTeacherAWDReviewContests(ctx context.Context, filter TeacherAWDReviewContestFilter) ([]assessmentdomain.TeacherAWDReviewContestCard, int64, TeacherAWDReviewContestSummary, error)
	FindTeacherAWDReviewContest(ctx context.Context, contestID int64) (*assessmentdomain.TeacherAWDReviewContestMeta, error)
	ListTeacherAWDReviewRounds(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewRoundSummary, error)
	ListTeacherAWDReviewTeams(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewTeamSummary, error)
	ListTeacherAWDReviewRoundServices(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewServiceRecord, error)
	ListTeacherAWDReviewRoundAttacks(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewAttackRecord, error)
	ListTeacherAWDReviewRoundTraffic(ctx context.Context, contestID, roundID int64) ([]assessmentdomain.TeacherAWDReviewTrafficRecord, error)
}
