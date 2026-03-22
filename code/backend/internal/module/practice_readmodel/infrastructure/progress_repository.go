package infrastructure

import (
	"context"
	"fmt"

	"ctf-platform/internal/model"
	readmodelapp "ctf-platform/internal/module/practice_readmodel/application"
)

func (r *Repository) GetUserProgress(ctx context.Context, userID int64) (totalScore int, totalSolved int, err error) {
	var result struct {
		TotalScore  int `gorm:"column:total_score"`
		TotalSolved int `gorm:"column:total_solved"`
	}
	err = r.dbWithContext(ctx).Table("submissions s").
		Select("COALESCE(SUM(c.points), 0) AS total_score, COUNT(DISTINCT s.challenge_id) AS total_solved").
		Joins("JOIN challenges c ON s.challenge_id = c.id").
		Where("s.user_id = ? AND s.is_correct = ? AND c.status = ?", userID, true, model.ChallengeStatusPublished).
		Scan(&result).Error
	if err != nil {
		return 0, 0, fmt.Errorf("get user progress: %w", err)
	}
	return result.TotalScore, result.TotalSolved, nil
}

func (r *Repository) GetCategoryStats(ctx context.Context, userID int64) ([]readmodelapp.CategoryProgressStat, error) {
	stats := make([]readmodelapp.CategoryProgressStat, 0)
	if err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.category,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved,
			COUNT(DISTINCT c.id) AS total
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ? AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.category
		ORDER BY c.category
	`, userID, model.ChallengeStatusPublished).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("get category stats: %w", err)
	}
	return stats, nil
}

func (r *Repository) GetDifficultyStats(ctx context.Context, userID int64) ([]readmodelapp.DifficultyProgressStat, error) {
	stats := make([]readmodelapp.DifficultyProgressStat, 0)
	if err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.difficulty,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved,
			COUNT(DISTINCT c.id) AS total
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ? AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.difficulty
		ORDER BY
			CASE c.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'insane' THEN 5
			END
	`, userID, model.ChallengeStatusPublished).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("get difficulty stats: %w", err)
	}
	return stats, nil
}

func (r *Repository) GetUserRank(ctx context.Context, userID int64) (int, error) {
	var rank int
	if err := r.dbWithContext(ctx).Raw(`
		WITH ranked_users AS (
			SELECT
				s.user_id,
				RANK() OVER (ORDER BY SUM(c.points) DESC) AS rank
			FROM submissions s
			JOIN challenges c ON s.challenge_id = c.id
			WHERE s.is_correct = TRUE AND c.status = ?
			GROUP BY s.user_id
		)
		SELECT COALESCE(rank, (SELECT COUNT(DISTINCT user_id) + 1 FROM ranked_users))
		FROM ranked_users WHERE user_id = ?
	`, model.ChallengeStatusPublished, userID).Scan(&rank).Error; err != nil {
		return 0, fmt.Errorf("get user rank: %w", err)
	}
	return rank, nil
}
