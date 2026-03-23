package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

type ChallengeRepository struct {
	db *gorm.DB
}

func NewChallengeRepository(db *gorm.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) AddChallenge(ctx context.Context, cc *model.ContestChallenge) error {
	return r.db.WithContext(ctx).Create(cc).Error
}

func (r *ChallengeRepository) RemoveChallenge(ctx context.Context, contestID, challengeID int64) error {
	return r.db.WithContext(ctx).Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Delete(&model.ContestChallenge{}).Error
}

func (r *ChallengeRepository) UpdateChallenge(ctx context.Context, contestID, challengeID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Updates(updates).Error
}

func (r *ChallengeRepository) ListChallenges(ctx context.Context, contestID int64, visibleOnly bool) ([]*model.ContestChallenge, error) {
	var challenges []*model.ContestChallenge
	query := r.db.WithContext(ctx).Where("contest_id = ?", contestID)
	if visibleOnly {
		query = query.Where("is_visible = ?", true)
	}
	err := query.
		Order("\"order\" ASC, created_at ASC").
		Find(&challenges).Error
	return challenges, err
}

func (r *ChallengeRepository) Exists(ctx context.Context, contestID, challengeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error
	return count > 0, err
}

func (r *ChallengeRepository) HasSubmissions(ctx context.Context, contestID, challengeID int64) (bool, error) {
	if !r.db.WithContext(ctx).Migrator().HasColumn(&model.Submission{}, "contest_id") {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&model.Submission{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error
	return count > 0, err
}
