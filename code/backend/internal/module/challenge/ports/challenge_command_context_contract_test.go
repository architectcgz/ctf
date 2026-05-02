package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeCommandRepository struct{}

func (ctxOnlyChallengeCommandRepository) CreateWithHints(context.Context, *model.Challenge, []*model.ChallengeHint) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) FindByID(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) Update(context.Context, *model.Challenge) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) UpdateWithHints(context.Context, *model.Challenge, []*model.ChallengeHint, bool) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) Delete(context.Context, int64) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) HasRunningInstances(context.Context, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeCommandRepository) CreatePublishCheckJob(context.Context, *model.ChallengePublishCheckJob) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) FindPublishCheckJobByID(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) FindActivePublishCheckJobByChallengeID(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) FindLatestPublishCheckJobByChallengeID(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) ListPendingPublishCheckJobs(context.Context, int) ([]*model.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) TryStartPublishCheckJob(context.Context, int64, time.Time) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeCommandRepository) UpdatePublishCheckJob(context.Context, *model.ChallengePublishCheckJob) error {
	return nil
}

var _ challengeports.ChallengeWriteRepository = (*ctxOnlyChallengeCommandRepository)(nil)
var _ challengeports.ChallengeInstanceUsageRepository = (*ctxOnlyChallengeCommandRepository)(nil)
var _ challengeports.ChallengePublishCheckRepository = (*ctxOnlyChallengeCommandRepository)(nil)
