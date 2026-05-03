package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (r *Repository) ApplyStatusTransition(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	return r.applyStatusTransitionWithUpdates(ctx, transition, map[string]any{
		"status":         transition.ToStatus,
		"status_version": transition.FromStatusVersion + 1,
		"updated_at":     transition.OccurredAt,
	})
}

func (r *Repository) UpdateContestWithStatusTransition(ctx context.Context, contest *model.Contest, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	if contest == nil {
		return contestdomain.ContestStatusTransitionResult{Transition: transition}, contestdomain.ErrContestNotFound
	}

	return r.applyStatusTransitionWithUpdates(ctx, transition, map[string]any{
		"title":          contest.Title,
		"description":    contest.Description,
		"mode":           contest.Mode,
		"start_time":     contest.StartTime,
		"end_time":       contest.EndTime,
		"freeze_time":    contest.FreezeTime,
		"status":         contest.Status,
		"status_version": transition.FromStatusVersion + 1,
		"updated_at":     transition.OccurredAt,
	})
}

func (r *Repository) applyStatusTransitionWithUpdates(ctx context.Context, transition contestdomain.ContestStatusTransition, updates map[string]any) (contestdomain.ContestStatusTransitionResult, error) {
	result := contestdomain.ContestStatusTransitionResult{Transition: transition}
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&model.Contest{}).
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
	err := db.Model(&model.Contest{}).
		Select("1").
		Where("id = ?", id).
		Limit(1).
		Find(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}
