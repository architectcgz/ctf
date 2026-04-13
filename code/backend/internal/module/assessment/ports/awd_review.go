package ports

import (
	"context"

	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

type TeacherAWDReviewRepository interface {
	ListTeacherAWDReviewContests(ctx context.Context) ([]assessmentdomain.TeacherAWDReviewContestCard, error)
	FindTeacherAWDReviewContest(ctx context.Context, contestID int64) (*assessmentdomain.TeacherAWDReviewContestMeta, error)
	ListTeacherAWDReviewRounds(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewRoundSummary, error)
	ListTeacherAWDReviewTeams(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewTeamSummary, error)
	ListTeacherAWDReviewRoundServices(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewServiceRecord, error)
	ListTeacherAWDReviewRoundAttacks(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewAttackRecord, error)
	ListTeacherAWDReviewRoundTraffic(ctx context.Context, contestID, roundID int64) ([]assessmentdomain.TeacherAWDReviewTrafficRecord, error)
}
