package ports_test

import (
	"context"
	"time"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type ctxOnlyChallengeCommandRepository struct{}

func (ctxOnlyChallengeCommandRepository) CreateWithHints(context.Context, *challengeports.ChallengeWriteModel, []*challengeentity.ChallengeHint) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) FindByID(context.Context, int64) (*challengeports.ChallengeWriteModel, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) LockChallengeByID(context.Context, int64) (*challengeports.ChallengeWriteModel, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) Update(context.Context, *challengeports.ChallengeWriteModel) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) MarkChallengePublished(context.Context, int64, time.Time) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) UpdateWithHints(context.Context, *challengeports.ChallengeWriteModel, []*challengeentity.ChallengeHint, bool) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) Delete(context.Context, int64) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) HasRunningInstances(context.Context, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeCommandRepository) CreatePublishCheckJob(context.Context, *challengeentity.ChallengePublishCheckJob) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) FindPublishCheckJobByID(context.Context, int64) (*challengeentity.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) FindActivePublishCheckJobByChallengeID(context.Context, int64) (*challengeentity.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) FindLatestPublishCheckJobByChallengeID(context.Context, int64) (*challengeentity.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) ListPendingPublishCheckJobs(context.Context, int) ([]*challengeentity.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (ctxOnlyChallengeCommandRepository) TryStartPublishCheckJob(context.Context, int64, time.Time) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeCommandRepository) UpdatePublishCheckJob(context.Context, *challengeentity.ChallengePublishCheckJob) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) WithinPublishCheckOutboxTx(context.Context, func(challengeports.ChallengePublishCheckOutboxTxRepository) error) error {
	return nil
}

func (ctxOnlyChallengeCommandRepository) EnqueueOutboxEvent(context.Context, platformevents.OutboxEvent) error {
	return nil
}

var _ challengeports.ChallengeWriteRepository = (*ctxOnlyChallengeCommandRepository)(nil)
var _ challengeports.ChallengeInstanceUsageRepository = (*ctxOnlyChallengeCommandRepository)(nil)
var _ challengeports.ChallengePublishCheckRepository = (*ctxOnlyChallengeCommandRepository)(nil)
var _ challengeports.ChallengePublishCheckOutboxTxRepository = (*ctxOnlyChallengeCommandRepository)(nil)
var _ challengeports.ChallengePublishCheckOutboxTxManager = (*ctxOnlyChallengeCommandRepository)(nil)
