package infrastructure

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

var _ contestports.ContestStatusSideEffectStore = (*ContestStatusSideEffectStore)(nil)

type ContestStatusSideEffectStore struct {
	cache *redislib.Client
}

func NewContestStatusSideEffectStore(cache *redislib.Client) *ContestStatusSideEffectStore {
	if cache == nil {
		return nil
	}
	return &ContestStatusSideEffectStore{cache: cache}
}

func (s *ContestStatusSideEffectStore) CreateFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil || contestID <= 0 {
		return nil
	}

	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	return s.cache.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err()
}

func (s *ContestStatusSideEffectStore) ClearFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil || contestID <= 0 {
		return nil
	}
	return s.cache.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err()
}

func (s *ContestStatusSideEffectStore) ClearEndedContestRuntimeState(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil || contestID <= 0 {
		return nil
	}
	return s.cache.Del(
		ctx,
		rediskeys.AWDCurrentRoundKey(contestID),
		rediskeys.AWDServiceStatusKey(contestID),
	).Err()
}
