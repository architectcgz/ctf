package ports_test

import (
	"context"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeImageUsageRepository struct{}

func (ctxOnlyChallengeImageUsageRepository) CountByImageIDWithContext(context.Context, int64) (int64, error) {
	return 0, nil
}

var _ challengeports.ChallengeImageUsageRepository = (*ctxOnlyChallengeImageUsageRepository)(nil)
