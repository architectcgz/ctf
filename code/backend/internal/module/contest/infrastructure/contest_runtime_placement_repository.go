package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
)

type ContestRuntimePlacementRepository struct {
	db *gorm.DB
}

func NewContestRuntimePlacementRepository(db *gorm.DB) *ContestRuntimePlacementRepository {
	return &ContestRuntimePlacementRepository{db: db}
}

func (r *ContestRuntimePlacementRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ContestRuntimePlacementRepository) FindActiveContestRuntimePlacement(ctx context.Context, contestID int64) (*contestentity.ContestRuntimePlacement, bool, error) {
	if r == nil || r.db == nil || contestID <= 0 {
		return nil, false, nil
	}

	var placement contestentity.ContestRuntimePlacement
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND status = ?", contestID, contestentity.ContestRuntimePlacementStatusActive).
		First(&placement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &placement, true, nil
}

func (r *ContestRuntimePlacementRepository) EnsureActiveContestRuntimePlacement(ctx context.Context, contestID, runtimeNodeID int64) (*contestentity.ContestRuntimePlacement, error) {
	if r == nil || r.db == nil || contestID <= 0 || runtimeNodeID <= 0 {
		return nil, nil
	}

	if placement, exists, err := r.FindActiveContestRuntimePlacement(ctx, contestID); err != nil || exists {
		return placement, err
	}

	now := time.Now().UTC()
	placement := &contestentity.ContestRuntimePlacement{
		ContestID:     contestID,
		RuntimeNodeID: runtimeNodeID,
		Status:        contestentity.ContestRuntimePlacementStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.dbWithContext(ctx).Create(placement).Error; err != nil {
		if existing, exists, findErr := r.FindActiveContestRuntimePlacement(ctx, contestID); findErr != nil {
			return nil, findErr
		} else if exists {
			return existing, nil
		}
		return nil, err
	}
	return placement, nil
}
