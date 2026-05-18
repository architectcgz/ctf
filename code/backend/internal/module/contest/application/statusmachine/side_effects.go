package statusmachine

import (
	"context"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type SideEffectRunner struct {
	store contestports.ContestStatusSideEffectStore
}

func NewSideEffectRunner(store contestports.ContestStatusSideEffectStore) *SideEffectRunner {
	return &SideEffectRunner{store: store}
}

func (r *SideEffectRunner) Run(ctx context.Context, result contestdomain.ContestStatusTransitionResult) error {
	if !result.Applied {
		return nil
	}
	if result.Transition.FromStatus == contestentity.ContestStatusRunning && result.Transition.ToStatus == contestentity.ContestStatusFrozen {
		return r.createFrozenSnapshot(ctx, result.Transition.ContestID)
	}
	if result.Transition.FromStatus == contestentity.ContestStatusFrozen && result.Transition.ToStatus == contestentity.ContestStatusRunning {
		return r.clearFrozenSnapshot(ctx, result.Transition.ContestID)
	}
	if result.Transition.ToStatus == contestentity.ContestStatusEnded {
		return r.clearEndedContestRuntimeState(ctx, result.Transition.ContestID)
	}
	return nil
}

func (r *SideEffectRunner) createFrozenSnapshot(ctx context.Context, contestID int64) error {
	if r == nil || r.store == nil {
		return nil
	}
	return r.store.CreateFrozenScoreboardSnapshot(ctx, contestID)
}

func (r *SideEffectRunner) clearEndedContestRuntimeState(ctx context.Context, contestID int64) error {
	if r == nil || r.store == nil {
		return nil
	}
	return r.store.ClearEndedContestRuntimeState(ctx, contestID)
}

func (r *SideEffectRunner) clearFrozenSnapshot(ctx context.Context, contestID int64) error {
	if r == nil || r.store == nil {
		return nil
	}
	return r.store.ClearFrozenScoreboardSnapshot(ctx, contestID)
}
