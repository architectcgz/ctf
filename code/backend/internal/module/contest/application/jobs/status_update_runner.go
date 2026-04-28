package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

func (u *StatusUpdater) updateStatuses(ctx context.Context) {
	lock, acquired, err := redislock.Acquire(ctx, u.redis, rediskeys.ContestStatusUpdateLockKey(), u.lockTTL)
	if err != nil {
		u.log.Error("acquire_contest_status_update_lock_failed", zap.Error(err))
		return
	}
	if !acquired {
		u.log.Debug("contest_status_update_lock_held_elsewhere")
		return
	}
	if lock != nil {
		defer func() {
			released, releaseErr := lock.Release(ctx)
			if releaseErr != nil {
				u.log.Error("release_contest_status_update_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(releaseErr))
				return
			}
			if !released {
				u.log.Warn("contest_status_update_lock_expired_before_release", zap.String("lock_key", lock.Key()))
			}
		}()
	}

	now := time.Now().UTC()
	statuses := []string{
		model.ContestStatusRegistration,
		model.ContestStatusRunning,
		model.ContestStatusFrozen,
	}

	contests, _, err := u.repo.ListByStatusesAndTimeRange(ctx, statuses, now, 0, u.batchSize)
	if err != nil {
		u.log.Error("list_contests_failed", zap.Error(err))
		return
	}

	for _, contest := range contests {
		newStatus := u.calculateStatus(contest, now)
		if newStatus != contest.Status {
			if u.shouldBlockAutomaticAWDStart(ctx, contest, newStatus) {
				continue
			}
			if contest.Status == model.ContestStatusRunning && newStatus == model.ContestStatusFrozen {
				u.createFrozenSnapshot(ctx, contest.ID)
			}
			if err := u.repo.UpdateStatus(ctx, contest.ID, newStatus); err != nil {
				u.log.Error("update_contest_status_failed", zap.Int64("contest_id", contest.ID), zap.Error(err))
			} else {
				if newStatus == model.ContestStatusEnded {
					u.clearEndedContestRuntimeState(ctx, contest.ID)
				}
				u.log.Info("contest_status_updated", zap.Int64("contest_id", contest.ID), zap.String("old_status", contest.Status), zap.String("new_status", newStatus))
			}
		}
	}
}
