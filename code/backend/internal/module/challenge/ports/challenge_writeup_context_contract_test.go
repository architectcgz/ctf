package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeWriteupRepository struct{}

func (ctxOnlyChallengeWriteupRepository) FindByIDWithContext(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindUserByIDWithContext(context.Context, int64) (*model.User, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindWriteupByChallengeIDWithContext(context.Context, int64) (*model.ChallengeWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) UpsertWriteupWithContext(context.Context, *model.ChallengeWriteup) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) DeleteWriteupByChallengeIDWithContext(context.Context, int64) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) FindReleasedWriteupByChallengeIDWithContext(context.Context, int64, time.Time) (*model.ChallengeWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) GetSolvedStatusWithContext(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeWriteupRepository) FindSubmissionWriteupByUserChallengeWithContext(context.Context, int64, int64) (*model.SubmissionWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindSubmissionWriteupByIDWithContext(context.Context, int64) (*model.SubmissionWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) UpsertSubmissionWriteupWithContext(context.Context, *model.SubmissionWriteup) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) GetTeacherSubmissionWriteupByIDWithContext(context.Context, int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) ListTeacherSubmissionWriteupsWithContext(context.Context, *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyChallengeWriteupRepository) ListRecommendedSolutionsByChallengeIDWithContext(context.Context, int64, time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) ListCommunitySolutionsByChallengeIDWithContext(context.Context, int64, *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	return nil, 0, nil
}

var _ challengeports.ChallengeWriteupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
