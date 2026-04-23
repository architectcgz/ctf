package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeFlagRepository struct{}

func (ctxOnlyChallengeFlagRepository) FindByIDWithContext(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeFlagRepository) UpdateWithContext(context.Context, *model.Challenge) error {
	return nil
}

var _ challengeports.ChallengeFlagRepository = (*ctxOnlyChallengeFlagRepository)(nil)
