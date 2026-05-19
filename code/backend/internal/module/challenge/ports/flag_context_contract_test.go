package ports_test

import (
	"context"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeFlagRepository struct{}

func (ctxOnlyChallengeFlagRepository) FindByID(context.Context, int64) (*challengeports.ChallengeFlag, error) {
	return nil, nil
}

func (ctxOnlyChallengeFlagRepository) Update(context.Context, *challengeports.ChallengeFlag) error {
	return nil
}

var _ challengeports.ChallengeFlagRepository = (*ctxOnlyChallengeFlagRepository)(nil)
