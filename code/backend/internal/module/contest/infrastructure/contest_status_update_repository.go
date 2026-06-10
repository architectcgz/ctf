package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *Repository) ApplyStatusTransition(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	return r.applyStatusTransitionWithUpdates(ctx, transition, map[string]any{
		"status":         transition.ToStatus,
		"status_version": transition.FromStatusVersion + 1,
		"updated_at":     transition.OccurredAt,
	})
}

func (r *Repository) UpdateContestWithStatusTransition(ctx context.Context, contest *contestentity.Contest, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	if contest == nil {
		return contestdomain.ContestStatusTransitionResult{Transition: transition}, contestdomain.ErrContestNotFound
	}

	return r.applyStatusTransitionWithUpdates(ctx, transition, contestStatusTransitionUpdates(contest, transition))
}

func (r *Repository) UpdateContestWithStatusTransitionAndRealtimeRelay(
	ctx context.Context,
	contest *contestentity.Contest,
	transition contestdomain.ContestStatusTransition,
	relay contestcontracts.RealtimeRelayEvent,
	dedupeKey string,
) (contestdomain.ContestStatusTransitionResult, error) {
	if contest == nil {
		return contestdomain.ContestStatusTransitionResult{Transition: transition}, contestdomain.ErrContestNotFound
	}

	return r.applyStatusTransitionWithUpdatesAndRealtimeRelay(ctx, transition, contestStatusTransitionUpdates(contest, transition), &relay, dedupeKey)
}

func (r *Repository) UpdateContestWithRealtimeRelay(ctx context.Context, contest *contestentity.Contest, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error {
	if contest == nil {
		return contestdomain.ErrContestNotFound
	}
	updatedAt := time.Now().UTC()
	contest.UpdatedAt = updatedAt

	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&contestentity.Contest{}).
			Where("id = ? AND deleted_at IS NULL", contest.ID).
			Updates(contestUpdateFields(contest, updatedAt))
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			return contestdomain.ErrContestNotFound
		}
		return NewRealtimeOutboxRepository(tx).EnqueueRealtimeRelay(ctx, relay, dedupeKey)
	})
}

func contestStatusTransitionUpdates(contest *contestentity.Contest, transition contestdomain.ContestStatusTransition) map[string]any {
	return map[string]any{
		"title":          contest.Title,
		"description":    contest.Description,
		"mode":           contest.Mode,
		"start_time":     contest.StartTime,
		"end_time":       contest.EndTime,
		"freeze_time":    contest.FreezeTime,
		"status":         contest.Status,
		"status_version": transition.FromStatusVersion + 1,
		"updated_at":     transition.OccurredAt,
	}
}

func contestUpdateFields(contest *contestentity.Contest, updatedAt time.Time) map[string]any {
	return map[string]any{
		"title":       contest.Title,
		"description": contest.Description,
		"mode":        contest.Mode,
		"start_time":  contest.StartTime,
		"end_time":    contest.EndTime,
		"freeze_time": contest.FreezeTime,
		"updated_at":  updatedAt,
	}
}

func (r *Repository) applyStatusTransitionWithUpdates(ctx context.Context, transition contestdomain.ContestStatusTransition, updates map[string]any) (contestdomain.ContestStatusTransitionResult, error) {
	return r.applyStatusTransitionWithUpdatesAndRealtimeRelay(ctx, transition, updates, nil, "")
}

func (r *Repository) applyStatusTransitionWithUpdatesAndRealtimeRelay(
	ctx context.Context,
	transition contestdomain.ContestStatusTransition,
	updates map[string]any,
	relay *contestcontracts.RealtimeRelayEvent,
	dedupeKey string,
) (contestdomain.ContestStatusTransitionResult, error) {
	result := contestdomain.ContestStatusTransitionResult{Transition: transition}
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&contestentity.Contest{}).
			Where("id = ? AND status = ? AND status_version = ? AND deleted_at IS NULL", transition.ContestID, transition.FromStatus, transition.FromStatusVersion).
			Updates(updates)
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			exists, existsErr := contestExistsTx(tx, transition.ContestID)
			if existsErr != nil {
				return existsErr
			}
			if !exists {
				return contestdomain.ErrContestNotFound
			}
			return nil
		}

		result.Applied = true
		result.StatusVersion = transition.FromStatusVersion + 1

		// The state row and transition journal must commit together; otherwise a side-effect replay worker
		// has no durable record to continue from after the API/job process exits.
		record, recordErr := upsertContestStatusTransitionRecord(tx, result)
		if recordErr != nil {
			return recordErr
		}
		result.RecordID = record.ID
		if relay != nil {
			if err := NewRealtimeOutboxRepository(tx).EnqueueRealtimeRelay(ctx, *relay, dedupeKey); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *Repository) contestExists(ctx context.Context, id int64) (bool, error) {
	return contestExistsTx(r.dbWithContext(ctx), id)
}

func contestExistsTx(db *gorm.DB, id int64) (bool, error) {
	var exists bool
	err := db.Model(&contestentity.Contest{}).
		Select("1").
		Where("id = ?", id).
		Limit(1).
		Find(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}
