package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *StatusUpdater) updateStatuses(ctx context.Context) {
	lock, acquired, err := u.acquireStatusUpdateLock(ctx)
	if err != nil {
		u.log.Error("acquire_contest_status_update_lock_failed", zap.Error(err))
		return
	}
	if !acquired {
		u.log.Debug("contest_status_update_lock_held_elsewhere")
		return
	}
	runCtx := ctx
	if lock != nil {
		// Status scanning can legitimately outlive the initial TTL, so the lock must be renewed for the whole run.
		var stopKeepalive func()
		runCtx, stopKeepalive = startRedisLockKeepalive(ctx, u.log, lock, redisLockKeepaliveConfig{
			Name: "contest_status_updater",
			TTL:  u.lockTTL,
		})
		defer func() {
			stopKeepalive()
			// Release should still run after the scheduler context is canceled so the lock is cleaned up on normal shutdown.
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer releaseCancel()
			released, releaseErr := lock.Release(releaseCtx)
			if releaseErr != nil {
				u.log.Error("release_contest_status_update_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(releaseErr))
				return
			}
			if !released {
				u.log.Warn("contest_status_update_lock_expired_before_release", zap.String("lock_key", lock.Key()))
			}
		}()
	}
	u.replayTransitionSideEffects(runCtx)

	now := time.Now().UTC()
	statuses := []string{
		model.ContestStatusRegistration,
		model.ContestStatusRunning,
		model.ContestStatusFrozen,
	}

	contests, _, err := u.repo.ListByStatusesAndTimeRange(runCtx, statuses, now, 0, u.batchSize)
	if err != nil {
		u.log.Error("list_contests_failed", zap.Error(err))
		return
	}

	for _, contest := range contests {
		if runCtx.Err() != nil {
			return
		}
		newStatus := u.calculateStatus(contest, now)
		if newStatus != contest.Status {
			if u.shouldBlockAutomaticAWDStart(runCtx, contest, newStatus) {
				continue
			}
			result, err := u.transitioner.Apply(runCtx, contestdomain.ContestStatusTransition{
				ContestID:         contest.ID,
				FromStatus:        contest.Status,
				ToStatus:          newStatus,
				FromStatusVersion: contest.StatusVersion,
				Reason:            contestdomain.ContestStatusTransitionReasonTimeWindow,
				OccurredAt:        now,
				AppliedBy:         contestStatusUpdaterAppliedBy,
			})
			if err != nil {
				u.log.Error("apply_contest_status_transition_failed",
					zap.Int64("contest_id", contest.ID),
					zap.String("from_status", contest.Status),
					zap.String("to_status", newStatus),
					zap.Error(err),
				)
				continue
			}
			if !result.Applied {
				continue
			}
			// Side effects belong to the instance that wins the compare-and-set transition; stale runners must skip them.
			if err := u.sideEffects.Run(runCtx, result); err != nil {
				if u.recorder != nil && result.RecordID > 0 {
					if recordErr := u.recorder.MarkTransitionSideEffectsFailed(runCtx, result.RecordID, err); recordErr != nil {
						u.log.Error("mark_contest_status_transition_failed",
							zap.Int64("contest_id", contest.ID),
							zap.Int64("transition_record_id", result.RecordID),
							zap.Error(recordErr),
						)
					}
				}
				u.log.Error("contest_status_transition_side_effects_failed",
					zap.Int64("contest_id", contest.ID),
					zap.String("from_status", contest.Status),
					zap.String("to_status", newStatus),
					zap.Error(err),
				)
				continue
			}
			if u.recorder != nil && result.RecordID > 0 {
				if recordErr := u.recorder.MarkTransitionSideEffectsSucceeded(runCtx, result.RecordID); recordErr != nil {
					u.log.Error("mark_contest_status_transition_succeeded_failed",
						zap.Int64("contest_id", contest.ID),
						zap.Int64("transition_record_id", result.RecordID),
						zap.Error(recordErr),
					)
				}
			}
			u.log.Info("contest_status_updated", zap.Int64("contest_id", contest.ID), zap.String("old_status", contest.Status), zap.String("new_status", newStatus))
		}
	}
}

func (u *StatusUpdater) acquireStatusUpdateLock(ctx context.Context) (contestports.ContestSchedulerLockLease, bool, error) {
	if u == nil || u.lockStore == nil || u.lockTTL <= 0 {
		return nil, true, nil
	}
	return u.lockStore.AcquireStatusUpdateLock(ctx, u.lockTTL)
}

func (u *StatusUpdater) replayTransitionSideEffects(ctx context.Context) {
	if u.replayer == nil || u.recorder == nil {
		return
	}

	results, err := u.replayer.ListTransitionsForSideEffectReplay(ctx, u.batchSize)
	if err != nil {
		u.log.Error("list_contest_status_transition_replays_failed", zap.Error(err))
		return
	}
	for _, result := range results {
		if ctx.Err() != nil {
			return
		}
		// Replay only re-runs idempotent cache/runtime cleanup. The DB transition already committed,
		// so this loop is responsible only for converging side effects after transient Redis/job failures.
		if err := u.sideEffects.Run(ctx, result); err != nil {
			if markErr := u.recorder.MarkTransitionSideEffectsFailed(ctx, result.RecordID, err); markErr != nil {
				u.log.Error("mark_replayed_contest_status_transition_failed",
					zap.Int64("contest_id", result.Transition.ContestID),
					zap.Int64("transition_record_id", result.RecordID),
					zap.Error(markErr),
				)
			}
			u.log.Error("replay_contest_status_transition_side_effects_failed",
				zap.Int64("contest_id", result.Transition.ContestID),
				zap.Int64("transition_record_id", result.RecordID),
				zap.Error(err),
			)
			continue
		}
		if err := u.recorder.MarkTransitionSideEffectsSucceeded(ctx, result.RecordID); err != nil {
			u.log.Error("mark_replayed_contest_status_transition_succeeded_failed",
				zap.Int64("contest_id", result.Transition.ContestID),
				zap.Int64("transition_record_id", result.RecordID),
				zap.Error(err),
			)
		}
	}
}
