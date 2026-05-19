package infrastructure

import (
	"context"

	"gorm.io/gorm/clause"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
)

func (r *Repository) FindChallengeScore(ctx context.Context, challengeID int64) (*practiceentity.Challenge, error) {
	var challenge practiceentity.Challenge
	if err := r.dbWithContext(ctx).
		Select("id, points, difficulty").
		Where("id = ?", challengeID).
		First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *Repository) FindChallengesScores(ctx context.Context, challengeIDs []int64) ([]practiceentity.Challenge, error) {
	if len(challengeIDs) == 0 {
		return []practiceentity.Challenge{}, nil
	}

	var challenges []practiceentity.Challenge
	err := r.dbWithContext(ctx).
		Select("id, points, difficulty").
		Where("id IN ?", challengeIDs).
		Find(&challenges).Error
	return challenges, err
}

func (r *Repository) ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error) {
	var challengeIDs []int64
	err := r.dbWithContext(ctx).
		Model(&submissionRow{}).
		Distinct("challenge_id").
		Where("user_id = ? AND is_correct = ?", userID, true).
		Pluck("challenge_id", &challengeIDs).Error
	return challengeIDs, err
}

func (r *Repository) UpsertUserScore(ctx context.Context, userScore *practiceentity.UserScore) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_score", "solved_count", "updated_at"}),
	}).Create(userScore).Error
}

func (r *Repository) FindUserScore(ctx context.Context, userID int64) (*practiceentity.UserScore, error) {
	var userScore practiceentity.UserScore
	if err := r.dbWithContext(ctx).
		Where("user_id = ?", userID).
		First(&userScore).Error; err != nil {
		return nil, err
	}
	return &userScore, nil
}

func (r *Repository) ListTopUserScores(ctx context.Context, limit int) ([]practiceentity.UserScore, error) {
	var scores []practiceentity.UserScore
	err := r.dbWithContext(ctx).
		Order("total_score DESC, updated_at ASC").
		Limit(limit).
		Find(&scores).Error
	return scores, err
}

func (r *Repository) FindUsersByIDs(ctx context.Context, userIDs []int64) ([]identitycontracts.User, error) {
	if len(userIDs) == 0 {
		return []identitycontracts.User{}, nil
	}

	var users []identitycontracts.User
	err := r.dbWithContext(ctx).
		Select("id, username, class_name").
		Where("id IN ?", userIDs).
		Find(&users).Error
	return users, err
}
