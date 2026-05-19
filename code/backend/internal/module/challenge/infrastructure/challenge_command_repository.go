package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeCommandRepositorySource interface {
	CreateWithHints(ctx context.Context, challenge *model.Challenge, hints []*challengeentity.ChallengeHint) error
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	Update(ctx context.Context, challenge *model.Challenge) error
	UpdateWithHints(ctx context.Context, challenge *model.Challenge, hints []*challengeentity.ChallengeHint, replaceHints bool) error
	Delete(ctx context.Context, id int64) error
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

func (r *ChallengeCommandRepository) CreateWithHints(ctx context.Context, challenge *challengeports.ChallengeWriteModel, hints []*challengeentity.ChallengeHint) error {
	rawChallenge := challengeWriteModelToModel(challenge)
	if err := r.source.CreateWithHints(ctx, rawChallenge, hints); err != nil {
		return err
	}
	if challenge != nil && rawChallenge != nil {
		*challenge = *challengeWriteModelFromModel(rawChallenge)
	}
	return nil
}

func (r *ChallengeCommandRepository) FindByID(ctx context.Context, id int64) (*challengeports.ChallengeWriteModel, error) {
	item, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeCommandChallengeNotFound
	}
	if err != nil || item == nil {
		return nil, err
	}
	return challengeWriteModelFromModel(item), nil
}

func (r *ChallengeCommandRepository) Update(ctx context.Context, challenge *challengeports.ChallengeWriteModel) error {
	return r.source.Update(ctx, challengeWriteModelToModel(challenge))
}

func (r *ChallengeCommandRepository) UpdateWithHints(ctx context.Context, challenge *challengeports.ChallengeWriteModel, hints []*challengeentity.ChallengeHint, replaceHints bool) error {
	return r.source.UpdateWithHints(ctx, challengeWriteModelToModel(challenge), hints, replaceHints)
}

func (r *ChallengeCommandRepository) Delete(ctx context.Context, id int64) error {
	return r.source.Delete(ctx, id)
}

func (r *ChallengeCommandRepository) HasRunningInstances(ctx context.Context, challengeID int64) (bool, error) {
	return r.source.HasRunningInstances(ctx, challengeID)
}

func (r *ChallengeCommandRepository) CreatePublishCheckJob(ctx context.Context, job *challengeentity.ChallengePublishCheckJob) error {
	return r.source.CreatePublishCheckJob(ctx, job)
}

func (r *ChallengeCommandRepository) FindPublishCheckJobByID(ctx context.Context, id int64) (*challengeentity.ChallengePublishCheckJob, error) {
	item, err := r.source.FindPublishCheckJobByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengePublishCheckJobNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) FindActivePublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengePublishCheckJob, error) {
	item, err := r.source.FindActivePublishCheckJobByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengePublishCheckJobNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) FindLatestPublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengePublishCheckJob, error) {
	item, err := r.source.FindLatestPublishCheckJobByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengePublishCheckJobNotFound
	}
	return item, err
}

func (r *ChallengeCommandRepository) ListPendingPublishCheckJobs(ctx context.Context, limit int) ([]*challengeentity.ChallengePublishCheckJob, error) {
	return r.source.ListPendingPublishCheckJobs(ctx, limit)
}

func (r *ChallengeCommandRepository) TryStartPublishCheckJob(ctx context.Context, id int64, startedAt time.Time) (bool, error) {
	return r.source.TryStartPublishCheckJob(ctx, id, startedAt)
}

func (r *ChallengeCommandRepository) UpdatePublishCheckJob(ctx context.Context, job *challengeentity.ChallengePublishCheckJob) error {
	return r.source.UpdatePublishCheckJob(ctx, job)
}

func challengeWriteModelFromModel(source *model.Challenge) *challengeports.ChallengeWriteModel {
	if source == nil {
		return nil
	}
	return &challengeports.ChallengeWriteModel{
		ID:              source.ID,
		PackageSlug:     source.PackageSlug,
		Title:           source.Title,
		Description:     source.Description,
		Category:        source.Category,
		Difficulty:      source.Difficulty,
		Points:          source.Points,
		ImageID:         source.ImageID,
		AttachmentURL:   source.AttachmentURL,
		Status:          string(source.Status),
		FlagType:        source.FlagType,
		FlagHash:        source.FlagHash,
		FlagSalt:        source.FlagSalt,
		FlagRegex:       source.FlagRegex,
		FlagPrefix:      source.FlagPrefix,
		InstanceSharing: string(source.InstanceSharing),
		TargetProtocol:  source.TargetProtocol,
		TargetPort:      source.TargetPort,
		CreatedBy:       source.CreatedBy,
		CreatedAt:       source.CreatedAt,
		UpdatedAt:       source.UpdatedAt,
	}
}

func challengeWriteModelToModel(source *challengeports.ChallengeWriteModel) *model.Challenge {
	if source == nil {
		return nil
	}
	return &model.Challenge{
		ID:              source.ID,
		PackageSlug:     source.PackageSlug,
		Title:           source.Title,
		Description:     source.Description,
		Category:        source.Category,
		Difficulty:      source.Difficulty,
		Points:          source.Points,
		ImageID:         source.ImageID,
		AttachmentURL:   source.AttachmentURL,
		Status:          model.ChallengeStatus(source.Status),
		FlagType:        source.FlagType,
		FlagHash:        source.FlagHash,
		FlagSalt:        source.FlagSalt,
		FlagRegex:       source.FlagRegex,
		FlagPrefix:      source.FlagPrefix,
		InstanceSharing: model.InstanceSharing(source.InstanceSharing),
		TargetProtocol:  source.TargetProtocol,
		TargetPort:      source.TargetPort,
		CreatedBy:       source.CreatedBy,
		CreatedAt:       source.CreatedAt,
		UpdatedAt:       source.UpdatedAt,
	}
}

var _ challengeports.ChallengeWriteRepository = (*ChallengeCommandRepository)(nil)
var _ challengeports.ChallengeInstanceUsageRepository = (*ChallengeCommandRepository)(nil)
var _ challengeports.ChallengePublishCheckRepository = (*ChallengeCommandRepository)(nil)
