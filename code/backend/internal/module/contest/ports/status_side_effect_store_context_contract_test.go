package ports_test

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyContestStatusSideEffectStore struct{}

func (ctxOnlyContestStatusSideEffectStore) CreateFrozenScoreboardSnapshot(context.Context, int64) error {
	return nil
}

func (ctxOnlyContestStatusSideEffectStore) ClearFrozenScoreboardSnapshot(context.Context, int64) error {
	return nil
}

func (ctxOnlyContestStatusSideEffectStore) ClearEndedContestRuntimeState(context.Context, int64) error {
	return nil
}

var _ contestports.ContestStatusSideEffectStore = (*ctxOnlyContestStatusSideEffectStore)(nil)
