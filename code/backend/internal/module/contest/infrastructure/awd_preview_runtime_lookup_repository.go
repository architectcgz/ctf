package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

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

func (r *AWDPreviewRuntimeChallengeRepository) FindAWDChallengeByID(ctx context.Context, id int64) (*challengecontracts.AWDChallenge, error) {
	challenge, err := r.source.FindAWDChallengeByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDPreviewChallengeNotFound
	}
	return challenge, err
}

func (r *AWDPreviewRuntimeChallengeRepository) ListAWDChallenges(ctx context.Context, query *challengecontracts.AWDChallengeQuery) ([]*challengecontracts.AWDChallenge, int64, error) {
	return r.source.ListAWDChallenges(ctx, query)
}

var _ challengeports.AWDChallengeQueryRepository = (*AWDPreviewRuntimeChallengeRepository)(nil)
