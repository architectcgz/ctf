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

func (ctxOnlyChallengeWriteupRepository) FindUserByID(context.Context, int64) (*model.User, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindWriteupByChallengeID(context.Context, int64) (*model.ChallengeWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) UpsertWriteup(context.Context, *model.ChallengeWriteup) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) DeleteWriteupByChallengeID(context.Context, int64) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) FindReleasedWriteupByChallengeID(context.Context, int64, time.Time) (*model.ChallengeWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) GetSolvedStatusWithContext(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeWriteupRepository) FindSubmissionWriteupByUserChallenge(context.Context, int64, int64) (*model.SubmissionWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) FindSubmissionWriteupByID(context.Context, int64) (*model.SubmissionWriteup, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) UpsertSubmissionWriteup(context.Context, *model.SubmissionWriteup) error {
	return nil
}

func (ctxOnlyChallengeWriteupRepository) GetTeacherSubmissionWriteupByID(context.Context, int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) ListTeacherSubmissionWriteups(context.Context, *dto.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyChallengeWriteupRepository) ListRecommendedSolutionsByChallengeID(context.Context, int64, time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	return nil, nil
}

func (ctxOnlyChallengeWriteupRepository) ListCommunitySolutionsByChallengeID(context.Context, int64, *dto.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	return nil, 0, nil
}

var _ challengeports.ChallengeWriteupRepository = (*ctxOnlyChallengeWriteupRepository)(nil)
