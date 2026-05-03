package infrastructure

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (r *Repository) RecordAppliedTransition(ctx context.Context, transitionResult contestdomain.ContestStatusTransitionResult) (int64, error) {
	record, err := upsertContestStatusTransitionRecord(r.dbWithContext(ctx), transitionResult)
	if err != nil {
		return 0, err
	}
	if record.ID <= 0 {
		return 0, fmt.Errorf("contest status transition record missing after upsert")
	}
	return record.ID, nil
}

func (r *Repository) ListTransitionsForSideEffectReplay(ctx context.Context, limit int) ([]contestdomain.ContestStatusTransitionResult, error) {
	query := r.dbWithContext(ctx).
		Model(&model.ContestStatusTransition{}).
		Where("side_effect_status IN ?", []string{
			contestdomain.ContestStatusTransitionSideEffectPending,
			contestdomain.ContestStatusTransitionSideEffectFailed,
		}).
		Order("occurred_at ASC, id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var records []model.ContestStatusTransition
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	results := make([]contestdomain.ContestStatusTransitionResult, 0, len(records))
	for _, record := range records {
		results = append(results, contestdomain.ContestStatusTransitionResult{
			Transition: contestdomain.ContestStatusTransition{
				ContestID:  record.ContestID,
				FromStatus: record.FromStatus,
				ToStatus:   record.ToStatus,
				Reason:     record.Reason,
				OccurredAt: record.OccurredAt,
				AppliedBy:  record.AppliedBy,
			},
			Applied:       true,
			StatusVersion: record.StatusVersion,
			RecordID:      record.ID,
		})
	}
	return results, nil
}

func upsertContestStatusTransitionRecord(db *gorm.DB, transitionResult contestdomain.ContestStatusTransitionResult) (*model.ContestStatusTransition, error) {
	record := model.ContestStatusTransition{
		ContestID:        transitionResult.Transition.ContestID,
		StatusVersion:    transitionResult.StatusVersion,
		FromStatus:       transitionResult.Transition.FromStatus,
		ToStatus:         transitionResult.Transition.ToStatus,
		Reason:           transitionResult.Transition.Reason,
		AppliedBy:        transitionResult.Transition.AppliedBy,
		SideEffectStatus: contestdomain.ContestStatusTransitionSideEffectPending,
		OccurredAt:       transitionResult.Transition.OccurredAt,
		CreatedAt:        transitionResult.Transition.OccurredAt,
		UpdatedAt:        transitionResult.Transition.OccurredAt,
	}

	dbResult := db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "contest_id"}, {Name: "status_version"}},
			DoNothing: true,
		}).
		Create(&record)
	if dbResult.Error != nil {
		return nil, dbResult.Error
	}
	if dbResult.RowsAffected == 0 {
		if err := db.
			Model(&model.ContestStatusTransition{}).
			Select("id").
			Where("contest_id = ? AND status_version = ?", record.ContestID, record.StatusVersion).
			First(&record).Error; err != nil {
			return nil, err
		}
	}
	return &record, nil
}

func (r *Repository) MarkTransitionSideEffectsSucceeded(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).Model(&model.ContestStatusTransition{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"side_effect_status": contestdomain.ContestStatusTransitionSideEffectSucceeded,
			"side_effect_error":  "",
		}).Error
}

func (r *Repository) MarkTransitionSideEffectsFailed(ctx context.Context, id int64, cause error) error {
	if id <= 0 {
		return nil
	}
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	return r.dbWithContext(ctx).Model(&model.ContestStatusTransition{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"side_effect_status": contestdomain.ContestStatusTransitionSideEffectFailed,
			"side_effect_error":  message,
		}).Error
}
