package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func applyAWDContestTeamScores(
	ctx context.Context,
	db *gorm.DB,
	teams []model.Team,
	serviceScoreMap map[int64]awdServiceScoreTotal,
	attackMap map[int64]awdAttackScoreRow,
) error {
	for _, team := range teams {
		serviceScore := serviceScoreMap[team.ID]
		attack := attackMap[team.ID]
		lastSolveAt := (*time.Time)(nil)
		if !attack.CreatedAt.IsZero() {
			lastSolveAt = &attack.CreatedAt
		}
		updates := map[string]any{
			"total_score":   serviceScore.SLAScore + serviceScore.DefenseScore + attack.ScoreGained,
			"last_solve_at": lastSolveAt,
		}
		if err := db.WithContext(ctx).
			Model(&model.Team{}).
			Where("id = ?", team.ID).
			Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}
