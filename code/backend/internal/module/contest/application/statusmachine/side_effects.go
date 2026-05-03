package statusmachine

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type SideEffectRunner struct {
	redis *redislib.Client
}

func NewSideEffectRunner(redis *redislib.Client) *SideEffectRunner {
	return &SideEffectRunner{redis: redis}
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
	if r.redis == nil {
		return nil
	}
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	return r.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err()
}

func (r *SideEffectRunner) clearEndedContestRuntimeState(ctx context.Context, contestID int64) error {
	if r.redis == nil || contestID <= 0 {
		return nil
	}
	return r.redis.Del(
		ctx,
		rediskeys.AWDCurrentRoundKey(contestID),
		rediskeys.AWDServiceStatusKey(contestID),
	).Err()
}

func (r *SideEffectRunner) clearFrozenSnapshot(ctx context.Context, contestID int64) error {
	if r.redis == nil || contestID <= 0 {
		return nil
	}
	return r.redis.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err()
}
