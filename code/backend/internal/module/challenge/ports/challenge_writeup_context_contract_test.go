package ports_test

import (
	"context"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type ctxOnlyChallengeWriteupRepository struct{}

func (ctxOnlyChallengeWriteupRepository) FindByID(context.Context, int64) (*challengeports.ChallengeWriteupChallenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindUserByID(context.Context, int64) (*identitycontracts.User, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindWriteupByChallengeID(context.Context, int64) (*challengeentity.ChallengeWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) UpsertWriteup(context.Context, *challengeentity.ChallengeWriteup) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) DeleteWriteupByChallengeID(context.Context, int64) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) FindReleasedWriteupByChallengeID(context.Context, int64, time.Time) (*challengeentity.ChallengeWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) GetSolvedStatus(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeWriteupRepository) FindSubmissionWriteupByUserChallenge(context.Context, int64, int64) (*challengeentity.SubmissionWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindSubmissionWriteupByID(context.Context, int64) (*challengeentity.SubmissionWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) UpsertSubmissionWriteup(context.Context, *challengeentity.SubmissionWriteup) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) GetTeacherSubmissionWriteupByID(context.Context, int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) ListTeacherSubmissionWriteups(context.Context, *challengecontracts.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyChallengeWriteupRepository) ListRecommendedSolutionsByChallengeID(context.Context, int64, time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) ListCommunitySolutionsByChallengeID(context.Context, int64, *challengecontracts.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	return nil, 0, nil
}

var _ challengeports.ChallengeWriteupChallengeLookupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeWriteupUserLookupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeAdminWriteupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeReleasedWriteupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeWriteupSolveStatusRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeSubmissionWriteupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeTeacherSubmissionWriteupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
var _ challengeports.ChallengeSolutionQueryRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
