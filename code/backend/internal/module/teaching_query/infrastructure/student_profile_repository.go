package infrastructure

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
)

type StudentProfileRepository struct {
	db *gorm.DB
}

func NewStudentProfileRepository(db *gorm.DB) *StudentProfileRepository {
	return &StudentProfileRepository{db: db}
}

func (r *StudentProfileRepository) CountPublishedChallenges(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("challenges").
		Where("status = ?", challengecontracts.ChallengeStatusPublished).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count published challenges: %w", err)
	}
	return count, nil
}

func (r *StudentProfileRepository) CountSolvedChallenges(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("submissions AS s").
		Joins("JOIN challenges c ON c.id = s.challenge_id").
		Where("s.user_id = ? AND s.is_correct = ? AND c.status = ?", userID, true, challengecontracts.ChallengeStatusPublished).
		Distinct("s.challenge_id").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count solved challenges: %w", err)
	}
	return count, nil
}

func (r *StudentProfileRepository) GetCategoryProgress(ctx context.Context, userID int64) ([]queryports.ProgressRow, error) {
	rows := make([]progressRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.category AS key,
			COUNT(DISTINCT c.id) AS total,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved
		FROM challenges c
		LEFT JOIN submissions s
			ON s.challenge_id = c.id
			AND s.user_id = ?
			AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.category
		ORDER BY c.category
	`, userID, challengecontracts.ChallengeStatusPublished).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get category progress: %w", err)
	}
	return toProgressRows(rows), nil
}

func (r *StudentProfileRepository) GetDifficultyProgress(ctx context.Context, userID int64) ([]queryports.ProgressRow, error) {
	rows := make([]progressRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.difficulty AS key,
			COUNT(DISTINCT c.id) AS total,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved
		FROM challenges c
		LEFT JOIN submissions s
			ON s.challenge_id = c.id
			AND s.user_id = ?
			AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.difficulty
		ORDER BY
			CASE c.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'hell' THEN 5
				ELSE 99
			END
	`, userID, challengecontracts.ChallengeStatusPublished).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get difficulty progress: %w", err)
	}
	return toProgressRows(rows), nil
}
