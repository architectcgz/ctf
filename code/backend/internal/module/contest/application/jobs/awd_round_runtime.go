package jobs

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func (u *AWDRoundUpdater) syncContestRounds(ctx context.Context, contest *model.Contest, now time.Time) {
	activeRound, totalRounds, ok := u.calculateRoundPlan(contest, now)
	if !ok {
		return
	}

	lockRound := activeRound
	if lockRound == 0 {
		lockRound = totalRounds
	}
	if lockRound <= 0 {
		return
	}

	acquired, err := u.acquireRoundLock(ctx, contest.ID, lockRound)
	if err != nil {
		if isExpectedShutdownError(err) {
			return
		}
		u.log.Error("acquire_awd_round_lock_failed", zap.Int64("contest_id", contest.ID), zap.Int("round_number", lockRound), zap.Error(err))
		return
	}
	if !acquired {
		return
	}

	if err := u.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		if isExpectedShutdownError(err) {
			return
		}
		u.log.Error("sync_awd_rounds_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Int("total_rounds", totalRounds), zap.Error(err))
		return
	}

	if err := u.syncRoundFlags(ctx, contest, activeRound, now); err != nil {
		if isExpectedShutdownError(err) {
			return
		}
		u.log.Error("sync_awd_round_flags_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Error(err))
	}
	if err := u.syncRoundServiceChecks(ctx, contest, activeRound); err != nil {
		if isExpectedShutdownError(err) {
			return
		}
		u.log.Error("sync_awd_service_checks_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Error(err))
	}
}

func (u *AWDRoundUpdater) EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	activeRound, totalRounds, ok := u.calculateRoundPlan(contest, now)
	if !ok || activeRound <= 0 {
		return gorm.ErrRecordNotFound
	}
	if err := u.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		return err
	}
	return u.syncRoundFlags(ctx, contest, activeRound, now)
}

func isExpectedShutdownError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
