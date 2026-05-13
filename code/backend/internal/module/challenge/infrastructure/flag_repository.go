package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type FlagRepository struct {
	source challengeports.ChallengeFlagRepository
}

func NewFlagRepository(source challengeports.ChallengeFlagRepository) *FlagRepository {
	if source == nil {
		return nil
	}
	return &FlagRepository{source: source}
}

func (r *FlagRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeFlagChallengeNotFound
	}
	return challenge, err
}

func (r *FlagRepository) Update(ctx context.Context, challenge *model.Challenge) error {
	return r.source.Update(ctx, challenge)
}

var _ challengeports.ChallengeFlagRepository = (*FlagRepository)(nil)
