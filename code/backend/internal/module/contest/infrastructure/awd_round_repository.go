package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
)

func (r *AWDRepository) CreateRound(ctx context.Context, round *model.AWDRound) error {
	return r.dbWithContext(ctx).Create(round).Error
}

func (r *AWDRepository) UpsertRound(ctx context.Context, round *model.AWDRound) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "contest_id"},
			{Name: "round_number"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"status":     round.Status,
			"started_at": round.StartedAt,
			"ended_at":   round.EndedAt,
			"updated_at": time.Now().UTC(),
		}),
	}).Create(round).Error
}

func (r *AWDRepository) ListRoundsByContest(ctx context.Context, contestID int64) ([]model.AWDRound, error) {
	var rounds []model.AWDRound
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("round_number ASC, id ASC").
		Find(&rounds).Error
	return rounds, err
}

func (r *AWDRepository) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("id = ? AND contest_id = ?", roundID, contestID).
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *AWDRepository) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND round_number = ?", contestID, roundNumber).
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *AWDRepository) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND status = ?", contestID, model.AWDRoundStatusRunning).
		Order("round_number DESC, id DESC").
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}
