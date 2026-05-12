package ports_test

import (
	"context"
	"time"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlySolvedCountCache struct{}

func (ctxOnlySolvedCountCache) GetSolvedCount(context.Context, int64) (int64, bool, error) {
	return 0, false, nil
}

func (ctxOnlySolvedCountCache) StoreSolvedCount(context.Context, int64, int64, time.Duration) error {
	return nil
}

var _ challengeports.ChallengeSolvedCountCache = (*ctxOnlySolvedCountCache)(nil)
