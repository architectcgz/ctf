package commands

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/apperror"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	platformevents "ctf-platform/internal/platform/events"
)

func (s *ScoreboardAdminService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return contestcontracts.ErrContestNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}

	now := time.Now().UTC()
	if contestdomain.ContestHasEndedAt(contest, now) {
		return contestcontracts.ErrContestEnded
	}

	effectiveEnd := contestdomain.ContestEffectiveEndTime(contest)
	storedFreezeTime := effectiveEnd.Add(-time.Duration(minutesBeforeEnd) * time.Minute).Add(-contestdomain.ContestPausedDuration(contest))
	contest.FreezeTime = &storedFreezeTime
	previousStatus := contest.Status
	previousVersion := contest.StatusVersion
	relay := scoreboardUpdatedRelay(contestID, now)
	dedupeKey := scoreboardOperationDedupeKey(contestID, "freeze", now)
	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	if !effectiveNow.Before(storedFreezeTime) {
		contest.Status = contestentity.ContestStatusFrozen
		contest.StatusVersion++
		contest.UpdatedAt = now
		if err := s.applyScoreboardStatusTransition(ctx, contest, previousStatus, previousVersion, relay, dedupeKey); err != nil {
			return err
		}
	} else {
		if err := s.updateScoreboardContest(ctx, contest, relay, dedupeKey); err != nil {
			return err
		}
	}

	if s.outbox == nil {
		return s.emitScoreboardUpdatedRealtime(ctx, contestID, "freeze", now)
	}
	return nil
}

func (s *ScoreboardAdminService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return contestcontracts.ErrContestNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}

	if contest.FreezeTime == nil && contest.Status != contestentity.ContestStatusFrozen {
		return contestcontracts.ErrScoreboardNotFrozen
	}

	contest.FreezeTime = nil
	previousStatus := contest.Status
	previousVersion := contest.StatusVersion
	now := time.Now().UTC()
	relay := scoreboardUpdatedRelay(contestID, now)
	dedupeKey := scoreboardOperationDedupeKey(contestID, "unfreeze", now)
	if contest.Status == contestentity.ContestStatusFrozen && !contestdomain.ContestHasEndedAt(contest, now) {
		contest.Status = contestentity.ContestStatusRunning
		contest.StatusVersion++
		contest.UpdatedAt = now
		if err := s.applyScoreboardStatusTransition(ctx, contest, previousStatus, previousVersion, relay, dedupeKey); err != nil {
			return err
		}
	} else {
		if err := s.updateScoreboardContest(ctx, contest, relay, dedupeKey); err != nil {
			return err
		}
	}

	if s.outbox == nil {
		return s.emitScoreboardUpdatedRealtime(ctx, contestID, "unfreeze", now)
	}
	return nil
}

func (s *ScoreboardAdminService) updateScoreboardContest(ctx context.Context, contest *contestentity.Contest, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error {
	if s.outbox != nil {
		if s.realtimeTx == nil {
			return apperror.ErrInternal.WithCause(fmt.Errorf("scoreboard realtime transaction repository unavailable"))
		}
		if err := s.realtimeTx.UpdateContestWithRealtimeRelay(ctx, contest, relay, dedupeKey); err != nil {
			if err == contestdomain.ErrContestNotFound {
				return contestcontracts.ErrContestNotFound
			}
			return apperror.ErrInternal.WithCause(err)
		}
		return nil
	}
	if err := s.repo.Update(ctx, contest); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ScoreboardAdminService) applyScoreboardStatusTransition(ctx context.Context, contest *contestentity.Contest, fromStatus string, fromVersion int64, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error {
	if contest == nil {
		return contestcontracts.ErrContestNotFound
	}
	if s.transition == nil {
		return apperror.ErrInternal.WithCause(fmt.Errorf("scoreboard transition repository unavailable"))
	}

	transition := contestdomain.ContestStatusTransition{
		ContestID:         contest.ID,
		FromStatus:        fromStatus,
		ToStatus:          contest.Status,
		FromStatusVersion: fromVersion,
		Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
		OccurredAt:        contest.UpdatedAt.UTC(),
		AppliedBy:         "scoreboard_admin_service",
	}
	var result contestdomain.ContestStatusTransitionResult
	var err error
	if s.outbox != nil {
		if s.realtimeTx == nil {
			return apperror.ErrInternal.WithCause(fmt.Errorf("scoreboard realtime transaction repository unavailable"))
		}
		result, err = s.realtimeTx.UpdateContestWithStatusTransitionAndRealtimeRelay(ctx, contest, transition, relay, dedupeKey)
	} else {
		result, err = s.transition.UpdateContestWithStatusTransition(ctx, contest, transition)
	}
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return contestcontracts.ErrContestNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}
	if !result.Applied {
		return apperror.ErrConflict.WithMessage("竞赛状态已变化，请刷新后重试")
	}
	contest.StatusVersion = result.StatusVersion

	// 封榜/解封同样会改写缓存快照，因此这里也必须留下可回放的 transition record。
	if err := s.sideEffects.Run(ctx, result); err != nil {
		if result.RecordID > 0 {
			if markErr := s.transition.MarkTransitionSideEffectsFailed(ctx, result.RecordID, err); markErr != nil {
				return apperror.ErrInternal.WithCause(fmt.Errorf("run scoreboard transition side effects: %w; mark failed: %v", err, markErr))
			}
		}
		return apperror.ErrInternal.WithCause(err)
	}
	if result.RecordID > 0 {
		if err := s.transition.MarkTransitionSideEffectsSucceeded(ctx, result.RecordID); err != nil {
			return apperror.ErrInternal.WithCause(err)
		}
	}
	return nil
}

func (s *ScoreboardAdminService) emitScoreboardUpdatedRealtime(ctx context.Context, contestID int64, operation string, occurredAt time.Time) error {
	if s == nil {
		return nil
	}
	if s.outbox != nil {
		return s.outbox.EnqueueRealtimeRelay(
			ctx,
			scoreboardUpdatedRelay(contestID, occurredAt),
			scoreboardOperationDedupeKey(contestID, operation, occurredAt),
		)
	}
	if s.eventBus == nil {
		return nil
	}
	return s.eventBus.Publish(ctx, platformevents.Event{
		Name:    contestcontracts.EventScoreboardUpdated,
		Payload: contestcontracts.ScoreboardUpdatedEvent{ContestID: contestID, OccurredAt: contestEventTimestamp(occurredAt)},
	})
}
