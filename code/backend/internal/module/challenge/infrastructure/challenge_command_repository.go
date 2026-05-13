package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeCommandRepositorySource interface {
	challengeports.ChallengeWriteRepository
	challengeports.ChallengeInstanceUsageRepository
	challengeports.ChallengePublishCheckRepository
}

type ChallengeCommandRepository struct {
	source challengeCommandRepositorySource
}

func NewChallengeCommandRepository(source challengeCommandRepositorySource) *ChallengeCommandRepository {
	if source == nil {
		return nil
	}
	return &ChallengeCommandRepository{source: source}
}

func (r *ChallengeCommandRepository) CreateWithHints(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error {
	return r.source.CreateWithHints(ctx, challenge, hints)
}

func (r *ChallengeCommandRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	item, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeCommandChallengeNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) Update(ctx context.Context, challenge *model.Challenge) error {
	return r.source.Update(ctx, challenge)
}

func (r *ChallengeCommandRepository) UpdateWithHints(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error {
	return r.source.UpdateWithHints(ctx, challenge, hints, replaceHints)
}

func (r *ChallengeCommandRepository) Delete(ctx context.Context, id int64) error {
	return r.source.Delete(ctx, id)
}

func (r *ChallengeCommandRepository) HasRunningInstances(ctx context.Context, challengeID int64) (bool, error) {
	return r.source.HasRunningInstances(ctx, challengeID)
}

func (r *ChallengeCommandRepository) CreatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error {
	return r.source.CreatePublishCheckJob(ctx, job)
}

func (r *ChallengeCommandRepository) FindPublishCheckJobByID(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error) {
	item, err := r.source.FindPublishCheckJobByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengePublishCheckJobNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) FindActivePublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
	item, err := r.source.FindActivePublishCheckJobByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengePublishCheckJobNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) FindLatestPublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
	item, err := r.source.FindLatestPublishCheckJobByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengePublishCheckJobNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) ListPendingPublishCheckJobs(ctx context.Context, limit int) ([]*model.ChallengePublishCheckJob, error) {
	return r.source.ListPendingPublishCheckJobs(ctx, limit)
}

func (r *ChallengeCommandRepository) TryStartPublishCheckJob(ctx context.Context, id int64, startedAt time.Time) (bool, error) {
	return r.source.TryStartPublishCheckJob(ctx, id, startedAt)
}

func (r *ChallengeCommandRepository) UpdatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error {
	return r.source.UpdatePublishCheckJob(ctx, job)
}

var _ challengeports.ChallengeWriteRepository = (*ChallengeCommandRepository)(nil)
var _ challengeports.ChallengeInstanceUsageRepository = (*ChallengeCommandRepository)(nil)
var _ challengeports.ChallengePublishCheckRepository = (*ChallengeCommandRepository)(nil)
