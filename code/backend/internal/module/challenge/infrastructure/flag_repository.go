package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type flagRepositorySource interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	Update(ctx context.Context, challenge *model.Challenge) error
}

type FlagRepository struct {
	source flagRepositorySource
}

func NewFlagRepository(source flagRepositorySource) *FlagRepository {
	if source == nil {
		return nil
	}
	return &FlagRepository{source: source}
}

func (r *FlagRepository) FindByID(ctx context.Context, id int64) (*challengeports.ChallengeFlag, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeFlagChallengeNotFound
	}
	if err != nil || challenge == nil {
		return nil, err
	}
	return &challengeports.ChallengeFlag{
		ID:              challenge.ID,
		FlagType:        challenge.FlagType,
		FlagPrefix:      challenge.FlagPrefix,
		FlagHash:        challenge.FlagHash,
		FlagSalt:        challenge.FlagSalt,
		FlagRegex:       challenge.FlagRegex,
		InstanceSharing: string(challenge.InstanceSharing),
	}, nil
}

func (r *FlagRepository) Update(ctx context.Context, challenge *challengeports.ChallengeFlag) error {
	if challenge == nil {
		return nil
	}
	existing, err := r.source.FindByID(ctx, challenge.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return challengeports.ErrChallengeFlagChallengeNotFound
	}
	if err != nil {
		return err
	}
	existing.FlagType = challenge.FlagType
	existing.FlagPrefix = challenge.FlagPrefix
	existing.FlagHash = challenge.FlagHash
	existing.FlagSalt = challenge.FlagSalt
	existing.FlagRegex = challenge.FlagRegex
	existing.InstanceSharing = model.InstanceSharing(challenge.InstanceSharing)
	return r.source.Update(ctx, existing)
}

var _ challengeports.ChallengeFlagRepository = (*FlagRepository)(nil)
