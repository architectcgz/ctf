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
	u.withSchedulerLock(ctx, func() {
		recentCutoff := now.Add(-u.cfg.RoundInterval)
		contests, err := u.repo.ListSchedulableAWDContests(ctx, now, recentCutoff, u.cfg.SchedulerBatchSize)
		if err != nil {
			u.log.Error("list_awd_contests_failed", zap.Error(err))
			return
		}

		for i := range contests {
			contestCopy := contests[i]
			u.syncContestRounds(ctx, &contestCopy, now)
		}
	})
}

func (u *AWDRoundUpdater) withSchedulerLock(ctx context.Context, fn func()) {
	lock, acquired, err := redislock.Acquire(ctx, u.redis, rediskeys.AWDSchedulerLockKey(), u.cfg.SchedulerLockTTL)
	if err != nil {
		u.log.Error("acquire_awd_scheduler_lock_failed", zap.Error(err))
		return
	}
	if !acquired {
		u.log.Debug("awd_scheduler_lock_held_elsewhere")
		return
	}
	if lock != nil {
		defer func() {
			released, releaseErr := lock.Release(ctx)
			if releaseErr != nil {
				u.log.Error("release_awd_scheduler_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(releaseErr))
				return
			}
			if !released {
				u.log.Warn("awd_scheduler_lock_expired_before_release", zap.String("lock_key", lock.Key()))
			}
		}()
	}
	fn()
}
