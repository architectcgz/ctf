package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type AWDChallengeRepository struct {
	source interface {
		challengeports.AWDChallengeCommandRepository
		challengeports.AWDChallengeQueryRepository
	}
}

func NewAWDChallengeRepository(source interface {
	challengeports.AWDChallengeCommandRepository
	challengeports.AWDChallengeQueryRepository
}) *AWDChallengeRepository {
	if source == nil {
		return nil
	}
	return &AWDChallengeRepository{source: source}
}

func (r *AWDChallengeRepository) CreateAWDChallenge(ctx context.Context, challenge *model.AWDChallenge) error {
	return r.source.CreateAWDChallenge(ctx, challenge)
}

func (r *AWDChallengeRepository) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	challenge, err := r.source.FindAWDChallengeByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrAWDChallengeNotFound
	}
	return challenge, err
}

func (r *AWDChallengeRepository) UpdateAWDChallenge(ctx context.Context, challenge *model.AWDChallenge) error {
	return r.source.UpdateAWDChallenge(ctx, challenge)
}

func (r *AWDChallengeRepository) DeleteAWDChallenge(ctx context.Context, id int64) error {
	return r.source.DeleteAWDChallenge(ctx, id)
}

func (r *AWDChallengeRepository) ListAWDChallenges(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return r.source.ListAWDChallenges(ctx, query)
}

var _ challengeports.AWDChallengeCommandRepository = (*AWDChallengeRepository)(nil)
var _ challengeports.AWDChallengeQueryRepository = (*AWDChallengeRepository)(nil)
