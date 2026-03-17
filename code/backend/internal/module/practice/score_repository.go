package practice

import (
	"context"

	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
)

func (r *Repository) FindChallengeScoreWithContext(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).
		Select("id, points, difficulty").
		Where("id = ?", challengeID).
		First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *Repository) FindChallengesScoresWithContext(ctx context.Context, challengeIDs []int64) ([]model.Challenge, error) {
	if len(challengeIDs) == 0 {
		return []model.Challenge{}, nil
	}

	var challenges []model.Challenge
	err := r.dbWithContext(ctx).
		Select("id, points, difficulty").
		Where("id IN ?", challengeIDs).
		Find(&challenges).Error
	return challenges, err
}

func (r *Repository) ListSolvedChallengeIDsWithContext(ctx context.Context, userID int64) ([]int64, error) {
	var challengeIDs []int64
	err := r.dbWithContext(ctx).
		Model(&model.Submission{}).
		Distinct("challenge_id").
		Where("user_id = ? AND is_correct = ?", userID, true).
		Pluck("challenge_id", &challengeIDs).Error
	return challengeIDs, err
}

func (r *Repository) UpsertUserScoreWithContext(ctx context.Context, userScore *model.UserScore) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_score", "solved_count", "updated_at"}),
	}).Create(userScore).Error
}

func (r *Repository) FindUserScoreWithContext(ctx context.Context, userID int64) (*model.UserScore, error) {
	var userScore model.UserScore
	if err := r.dbWithContext(ctx).
		Where("user_id = ?", userID).
		First(&userScore).Error; err != nil {
		return nil, err
	}
	return &userScore, nil
}

func (r *Repository) ListTopUserScoresWithContext(ctx context.Context, limit int) ([]model.UserScore, error) {
	var scores []model.UserScore
	err := r.dbWithContext(ctx).
		Order("total_score DESC, updated_at ASC").
		Limit(limit).
		Find(&scores).Error
	return scores, err
}

func (r *Repository) FindUsersByIDsWithContext(ctx context.Context, userIDs []int64) ([]model.User, error) {
	if len(userIDs) == 0 {
		return []model.User{}, nil
	}

	var users []model.User
	err := r.dbWithContext(ctx).
		Select("id, username").
		Where("id IN ?", userIDs).
		Find(&users).Error
	return users, err
}
