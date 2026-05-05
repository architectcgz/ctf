package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

func (u *AWDRoundUpdater) UpdateRoundsAt(ctx context.Context, now time.Time) {
	if u.repo == nil {
		return
	}
	now = now.UTC()
	u.withSchedulerLock(ctx, func(runCtx context.Context) {
		recentCutoff := now.Add(-u.cfg.RoundInterval)
		contests, err := u.repo.ListSchedulableAWDContests(runCtx, now, recentCutoff, u.cfg.SchedulerBatchSize)
		if err != nil {
			u.log.Error("list_awd_contests_failed", zap.Error(err))
			return
		}

		for i := range contests {
			if runCtx.Err() != nil {
				return
			}
			contestCopy := contests[i]
			u.syncContestRounds(runCtx, &contestCopy, now)
		}
	})
}

func (u *AWDRoundUpdater) withSchedulerLock(ctx context.Context, fn func(context.Context)) {
	lock, acquired, err := redislock.Acquire(ctx, u.redis, rediskeys.AWDSchedulerLockKey(), u.cfg.SchedulerLockTTL)
	if err != nil {
		u.log.Error("acquire_awd_scheduler_lock_failed", zap.Error(err))
		return
	}
	if !acquired {
		u.log.Debug("awd_scheduler_lock_held_elsewhere")
		return
	}
	runCtx := ctx
	if lock != nil {
		lockCtx, cancel := context.WithCancel(ctx)
		// 调度执行可能超过锁 TTL，持锁期间需要持续续租；一旦续租失败且确认失锁，当前运行应立即收敛。
		stopKeepalive := u.startSchedulerLockKeepalive(lockCtx, cancel, lock)
		runCtx = lockCtx
		defer func() {
			stopKeepalive()
			// 释放锁不应被本轮运行上下文的取消打断，否则会把“正常退出”放大成额外噪音。
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer releaseCancel()
			released, releaseErr := lock.Release(releaseCtx)
			if releaseErr != nil {
				u.log.Error("release_awd_scheduler_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(releaseErr))
				return
			}
			if !released {
				u.log.Warn("awd_scheduler_lock_expired_before_release", zap.String("lock_key", lock.Key()))
			}
		}()
	}
	fn(runCtx)
}

func (u *AWDRoundUpdater) startSchedulerLockKeepalive(ctx context.Context, cancel context.CancelFunc, lock *redislock.Lock) func() {
	interval := schedulerLockRefreshInterval(u.cfg.SchedulerLockTTL)
	keepaliveCtx, keepaliveCancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer close(done)
		defer ticker.Stop()

		for {
			select {
			case <-keepaliveCtx.Done():
				return
			case <-ticker.C:
				refreshed, err := lock.Refresh(keepaliveCtx, u.cfg.SchedulerLockTTL)
				if err != nil {
					if keepaliveCtx.Err() == nil {
						u.log.Error("refresh_awd_scheduler_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(err))
					}
					continue
				}
				if refreshed {
					continue
				}
				// Refresh 返回 false 说明 Redis 中的 token 已经不是自己，继续推进会破坏单实例调度约束。
				u.log.Warn("awd_scheduler_lock_lost_during_run", zap.String("lock_key", lock.Key()))
				cancel()
				return
			}
		}
	}()

	return func() {
		keepaliveCancel()
		<-done
	}
}

func schedulerLockRefreshInterval(ttl time.Duration) time.Duration {
	var interval time.Duration
	switch {
	case ttl <= 3*time.Second:
		interval = ttl / 2
	default:
		interval = ttl / 3
	}
	if interval <= 0 {
		interval = ttl
	}
	if interval <= 0 {
		interval = time.Millisecond
	}
	return interval
}
