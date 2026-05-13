package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
)

type AWDPreviewRuntimeChallengeRepository struct {
	source challengeports.AWDChallengeQueryRepository
}

func NewAWDPreviewRuntimeChallengeRepository(source challengeports.AWDChallengeQueryRepository) *AWDPreviewRuntimeChallengeRepository {
	if source == nil {
		return nil
	}
	return &AWDPreviewRuntimeChallengeRepository{source: source}
}

func (r *AWDPreviewRuntimeChallengeRepository) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	challenge, err := r.source.FindAWDChallengeByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDPreviewChallengeNotFound
	}
	return challenge, err
}

func (r *AWDPreviewRuntimeChallengeRepository) ListAWDChallenges(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return r.source.ListAWDChallenges(ctx, query)
}

type AWDPreviewRuntimeImageRepository struct {
	source challengecontracts.ImageStore
}

func NewAWDPreviewRuntimeImageRepository(source challengecontracts.ImageStore) *AWDPreviewRuntimeImageRepository {
	if source == nil {
		return nil
	}
	return &AWDPreviewRuntimeImageRepository{source: source}
}

func (r *AWDPreviewRuntimeImageRepository) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	image, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDPreviewImageNotFound
	}
	return image, err
}

var _ challengeports.AWDChallengeQueryRepository = (*AWDPreviewRuntimeChallengeRepository)(nil)
var _ challengecontracts.ImageStore = (*AWDPreviewRuntimeImageRepository)(nil)
