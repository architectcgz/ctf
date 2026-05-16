package commands

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/model"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardAdminService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now().UTC()
	if contestdomain.ContestHasEndedAt(contest, now) {
		return errcode.ErrContestEnded
	}

	effectiveEnd := contestdomain.ContestEffectiveEndTime(contest)
	storedFreezeTime := effectiveEnd.Add(-time.Duration(minutesBeforeEnd) * time.Minute).Add(-contestdomain.ContestPausedDuration(contest))
	contest.FreezeTime = &storedFreezeTime
	previousStatus := contest.Status
	previousVersion := contest.StatusVersion
	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	if !effectiveNow.Before(storedFreezeTime) {
		contest.Status = model.ContestStatusFrozen
		contest.StatusVersion++
		contest.UpdatedAt = now
		if err := s.applyScoreboardStatusTransition(ctx, contest, previousStatus, previousVersion); err != nil {
			return err
		}
	} else if err := s.repo.Update(ctx, contest); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	publishContestWeakEvent(ctx, s.eventBus, platformevents.Event{
		Name: contestcontracts.EventScoreboardUpdated,
		Payload: contestcontracts.ScoreboardUpdatedEvent{
			ContestID:  contestID,
			OccurredAt: contestEventTimestamp(now),
		},
	})
	return nil
}

func (s *ScoreboardAdminService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if contest.FreezeTime == nil && contest.Status != model.ContestStatusFrozen {
		return errcode.ErrScoreboardNotFrozen
	}

	contest.FreezeTime = nil
	previousStatus := contest.Status
	previousVersion := contest.StatusVersion
	now := time.Now().UTC()
	if contest.Status == model.ContestStatusFrozen && !contestdomain.ContestHasEndedAt(contest, now) {
		contest.Status = model.ContestStatusRunning
		contest.StatusVersion++
		contest.UpdatedAt = now
		if err := s.applyScoreboardStatusTransition(ctx, contest, previousStatus, previousVersion); err != nil {
			return err
		}
	} else if err := s.repo.Update(ctx, contest); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	publishContestWeakEvent(ctx, s.eventBus, platformevents.Event{
		Name: contestcontracts.EventScoreboardUpdated,
		Payload: contestcontracts.ScoreboardUpdatedEvent{
			ContestID:  contestID,
			OccurredAt: contestEventTimestamp(now),
		},
	})
	return nil
}

func (s *ScoreboardAdminService) applyScoreboardStatusTransition(ctx context.Context, contest *model.Contest, fromStatus string, fromVersion int64) error {
	if contest == nil {
		return errcode.ErrContestNotFound
	}
	if s.transition == nil {
		return errcode.ErrInternal.WithCause(fmt.Errorf("scoreboard transition repository unavailable"))
	}

	result, err := s.transition.UpdateContestWithStatusTransition(ctx, contest, contestdomain.ContestStatusTransition{
		ContestID:         contest.ID,
		FromStatus:        fromStatus,
		ToStatus:          contest.Status,
		FromStatusVersion: fromVersion,
		Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
		OccurredAt:        contest.UpdatedAt.UTC(),
		AppliedBy:         "scoreboard_admin_service",
	})
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if !result.Applied {
		return errcode.New(errcode.ErrConflict.Code, "竞赛状态已变化，请刷新后重试", errcode.ErrConflict.HTTPStatus)
	}
	contest.StatusVersion = result.StatusVersion

	// 封榜/解封同样会改写缓存快照，因此这里也必须留下可回放的 transition record。
	if err := s.sideEffects.Run(ctx, result); err != nil {
		if result.RecordID > 0 {
			if markErr := s.transition.MarkTransitionSideEffectsFailed(ctx, result.RecordID, err); markErr != nil {
				return errcode.ErrInternal.WithCause(fmt.Errorf("run scoreboard transition side effects: %w; mark failed: %v", err, markErr))
			}
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if result.RecordID > 0 {
		if err := s.transition.MarkTransitionSideEffectsSucceeded(ctx, result.RecordID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	return nil
}
