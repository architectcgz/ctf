package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
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
	lock, acquired, err := u.acquireSchedulerLock(ctx)
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
		var stopKeepalive func()
		runCtx, stopKeepalive = startRedisLockKeepalive(ctx, u.log, lock, redisLockKeepaliveConfig{
			Name: "awd_round_updater",
			TTL:  u.cfg.SchedulerLockTTL,
		})
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

func (u *AWDRoundUpdater) acquireSchedulerLock(ctx context.Context) (contestports.ContestSchedulerLockLease, bool, error) {
	if u == nil || u.stateStore == nil || u.cfg.SchedulerLockTTL <= 0 {
		return nil, true, nil
	}
	return u.stateStore.AcquireAWDSchedulerLock(ctx, u.cfg.SchedulerLockTTL)
}
