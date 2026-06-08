package infrastructure

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/infrastructure/redislock"
	"ctf-platform/internal/module/practice/infrastructure/cachekeys"
	practiceports "ctf-platform/internal/module/practice/ports"
)

var _ practiceports.PracticeInstanceSchedulerLockStore = (*SchedulerStateStore)(nil)
var _ practiceports.PracticeSchedulerLockLease = (*redislock.Lock)(nil)

type SchedulerStateStore struct {
	cache *redislib.Client
}

func NewSchedulerStateStore(cache *redislib.Client) *SchedulerStateStore {
	if cache == nil {
		return nil
	}
	return &SchedulerStateStore{cache: cache}
}

func (s *SchedulerStateStore) AcquireProvisioningSchedulerLock(ctx context.Context, ttl time.Duration) (practiceports.PracticeSchedulerLockLease, bool, error) {
	if s == nil || s.cache == nil {
		return nil, true, nil
	}
	return redislock.Acquire(ctx, s.cache, cachekeys.ProvisioningSchedulerLockKey(), ttl)
}
