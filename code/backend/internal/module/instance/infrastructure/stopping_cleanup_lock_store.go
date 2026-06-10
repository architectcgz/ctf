package infrastructure

import (
	"context"
	"errors"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/infrastructure/redislock"
	"ctf-platform/internal/module/instance/infrastructure/cachekeys"
	"ctf-platform/internal/shared/lockkeepalive"
)

type StoppingCleanupLockStore struct {
	cache   *redislib.Client
	lockTTL time.Duration
	logger  *zap.Logger
}

func NewStoppingCleanupLockStore(cache *redislib.Client, lockTTL time.Duration, logger *zap.Logger) *StoppingCleanupLockStore {
	if lockTTL <= 0 {
		lockTTL = 2 * time.Minute
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StoppingCleanupLockStore{
		cache:   cache,
		lockTTL: lockTTL,
		logger:  logger,
	}
}

func (s *StoppingCleanupLockStore) WithStoppingCleanupLock(ctx context.Context, fn func(context.Context)) (bool, error) {
	if ctx == nil {
		return false, errors.New("stopping cleanup lock requires context")
	}
	lock, acquired, err := redislock.Acquire(ctx, s.cache, cachekeys.StoppingCleanupLockKey(), s.lockTTL)
	if err != nil {
		return false, err
	}
	if !acquired {
		return false, nil
	}
	runCtx := ctx
	var stopKeepalive func()
	if lock != nil {
		runCtx, stopKeepalive = lockkeepalive.Start(ctx, s.logger, lock, "instance_stopping_cleanup", s.lockTTL)
	} else {
		stopKeepalive = func() {}
	}
	defer func() {
		stopKeepalive()
		if lock == nil {
			return
		}
		releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer releaseCancel()
		released, releaseErr := lock.Release(releaseCtx)
		if releaseErr != nil {
			if !errors.Is(releaseErr, context.Canceled) {
				s.logger.Error("释放 stopping 实例清理锁失败", zap.String("lock_key", lock.Key(releaseCtx)), zap.Error(releaseErr))
			}
			return
		}
		if !released && ctx.Err() == nil {
			s.logger.Warn("stopping 实例清理锁已过期或被覆盖", zap.String("lock_key", lock.Key(releaseCtx)))
		}
	}()

	if fn != nil {
		fn(runCtx)
	}
	return true, nil
}
