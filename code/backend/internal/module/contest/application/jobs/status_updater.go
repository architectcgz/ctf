package jobs

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

type StatusUpdater struct {
	repo      contestports.Repository
	redis     *redislib.Client
	log       *zap.Logger
	interval  time.Duration
	batchSize int
	lockTTL   time.Duration
}

func NewStatusUpdater(repo contestports.Repository, redis *redislib.Client, interval time.Duration, batchSize int, lockTTL time.Duration, log *zap.Logger) *StatusUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	if lockTTL <= 0 {
		lockTTL = 30 * time.Second
	}
	return &StatusUpdater{
		repo:      repo,
		redis:     redis,
		log:       log,
		interval:  interval,
		batchSize: batchSize,
		lockTTL:   lockTTL,
	}
}

func (u *StatusUpdater) Start(ctx context.Context) {
	u.updateStatuses(ctx)

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.updateStatuses(ctx)
		}
	}
}

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

	now := time.Now()

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

func (u *StatusUpdater) calculateStatus(contest *model.Contest, now time.Time) string {
	if contest.Status == model.ContestStatusDraft {
		return model.ContestStatusDraft
	}

	if now.Before(contest.StartTime) {
		return model.ContestStatusRegistration
	}

	if !now.Before(contest.EndTime) {
		return model.ContestStatusEnded
	}

	if contest.FreezeTime != nil && !now.Before(*contest.FreezeTime) {
		return model.ContestStatusFrozen
	}

	return model.ContestStatusRunning
}

func (u *StatusUpdater) createFrozenSnapshot(ctx context.Context, contestID int64) {
	if u.redis == nil {
		return
	}
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	if err := u.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err(); err != nil {
		u.log.Error("create_frozen_snapshot_failed", zap.Int64("contest_id", contestID), zap.Error(err))
	}
}

func (u *StatusUpdater) clearEndedContestRuntimeState(ctx context.Context, contestID int64) {
	if u.redis == nil || contestID <= 0 {
		return
	}
	if err := u.redis.Del(
		ctx,
		rediskeys.AWDCurrentRoundKey(contestID),
		rediskeys.AWDServiceStatusKey(contestID),
	).Err(); err != nil {
		u.log.Error("clear_ended_contest_runtime_state_failed", zap.Int64("contest_id", contestID), zap.Error(err))
	}
}
