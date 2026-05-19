package infrastructure

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
)

type challengeProjectionRow struct {
	ID         int64  `gorm:"column:id"`
	Title      string `gorm:"column:title"`
	Category   string `gorm:"column:category"`
	Difficulty string `gorm:"column:difficulty"`
	Points     int    `gorm:"column:points"`
	Status     string `gorm:"column:status"`
	FlagType   string `gorm:"column:flag_type"`
	FlagPrefix string `gorm:"column:flag_prefix"`
}

func (r *SubmissionRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error) {
	var registration contestentity.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *SubmissionRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error) {
	var contestChallenge contestentity.ContestChallenge
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&contestChallenge).Error; err != nil {
		return nil, err
	}
	return &contestChallenge, nil
}

func (r *SubmissionRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*contestentity.Challenge, error) {
	var row challengeProjectionRow
	if err := r.dbWithContext(ctx).
		Table("challenges").
		Select("id", "title", "category", "difficulty", "points", "status", "flag_type", "flag_prefix").
		Where("id = ?", challengeID).
		Where("deleted_at IS NULL").
		Take(&row).Error; err != nil {
		return nil, err
	}
	return &contestentity.Challenge{
		ID:         row.ID,
		Title:      row.Title,
		Category:   row.Category,
		Difficulty: row.Difficulty,
		Points:     row.Points,
		Status:     row.Status,
		FlagType:   row.FlagType,
		FlagPrefix: row.FlagPrefix,
	}, nil
}
