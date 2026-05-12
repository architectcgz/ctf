package statusmachine

import (
	"context"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
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
	if result.Transition.FromStatus == model.ContestStatusRunning && result.Transition.ToStatus == model.ContestStatusFrozen {
		return r.createFrozenSnapshot(ctx, result.Transition.ContestID)
	}
	if result.Transition.FromStatus == model.ContestStatusFrozen && result.Transition.ToStatus == model.ContestStatusRunning {
		return r.clearFrozenSnapshot(ctx, result.Transition.ContestID)
	}
	if result.Transition.ToStatus == model.ContestStatusEnded {
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
