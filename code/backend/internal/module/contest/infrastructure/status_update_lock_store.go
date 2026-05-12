package infrastructure

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

var _ contestports.ContestStatusUpdateLockStore = (*ContestStatusUpdateLockStore)(nil)

type ContestStatusUpdateLockStore struct {
	cache *redislib.Client
}

func NewContestStatusUpdateLockStore(cache *redislib.Client) *ContestStatusUpdateLockStore {
	if cache == nil {
		return nil
	}
	return &ContestStatusUpdateLockStore{cache: cache}
}

func (s *ContestStatusUpdateLockStore) AcquireStatusUpdateLock(ctx context.Context, ttl time.Duration) (contestports.ContestStatusUpdateLockLease, bool, error) {
	if s == nil || s.cache == nil || ttl <= 0 {
		return nil, true, nil
	}
	return redislock.Acquire(ctx, s.cache, rediskeys.ContestStatusUpdateLockKey(), ttl)
}
