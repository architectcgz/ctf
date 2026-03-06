package contest

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type StatusUpdater struct {
	repo      Repository
	redis     *redislib.Client
	log       *zap.Logger
	interval  time.Duration
	batchSize int
}

func NewStatusUpdater(repo Repository, redis *redislib.Client, interval time.Duration, batchSize int, log *zap.Logger) *StatusUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	return &StatusUpdater{
		repo:      repo,
		redis:     redis,
		log:       log,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (u *StatusUpdater) Start(ctx context.Context) {
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
