package contest

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
)

type StatusUpdater struct {
	repo      Repository
	log       *zap.Logger
	interval  time.Duration
	batchSize int
}

func NewStatusUpdater(repo Repository, interval time.Duration, batchSize int, log *zap.Logger) *StatusUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	return &StatusUpdater{
		repo:      repo,
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

	if now.After(contest.EndTime) {
		return model.ContestStatusEnded
	}

	return model.ContestStatusRunning
}
