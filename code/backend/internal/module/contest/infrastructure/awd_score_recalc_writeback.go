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
	defenseMap map[int64]int,
	attackMap map[int64]awdAttackScoreRow,
) error {
	for _, team := range teams {
		attack := attackMap[team.ID]
		lastSolveAt := (*time.Time)(nil)
		if !attack.CreatedAt.IsZero() {
			lastSolveAt = &attack.CreatedAt
		}
		updates := map[string]any{
			"total_score":   defenseMap[team.ID] + attack.ScoreGained,
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
