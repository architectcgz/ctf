package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func RecalculateAWDContestTeamScores(ctx context.Context, db *gorm.DB, contestID int64) error {
	if db == nil || contestID <= 0 {
		return nil
	}

	var teams []model.Team
	if err := db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error; err != nil {
		return err
	}
	if len(teams) == 0 {
		return nil
	}

	var serviceRows []awdServiceScoreRow
	if err := db.WithContext(ctx).
		Table("awd_team_services AS ats").
		Select("ats.team_id AS team_id, ats.defense_score AS defense_score, ats.check_result AS check_result").
		Joins("JOIN awd_rounds AS ar ON ar.id = ats.round_id").
		Where("ar.contest_id = ?", contestID).
		Scan(&serviceRows).Error; err != nil {
		return err
	}

	var attackRows []awdAttackScoreRow
	if err := db.WithContext(ctx).
		Table("awd_attack_logs AS aal").
		Select("aal.attacker_team_id AS team_id, aal.score_gained AS score_gained, aal.source AS source, aal.created_at AS created_at").
		Joins("JOIN awd_rounds AS ar ON ar.id = aal.round_id").
		Where("ar.contest_id = ? AND aal.score_gained > 0", contestID).
		Scan(&attackRows).Error; err != nil {
		return err
	}

	defenseMap := accumulateAWDDefenseScores(serviceRows)
	attackMap := accumulateAWDAttackScores(attackRows)

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

func SyncAWDContestScores(ctx context.Context, db *gorm.DB, redis redisScoreboardCache, contestID int64) error {
	if err := RecalculateAWDContestTeamScores(ctx, db, contestID); err != nil {
		return err
	}
	return RebuildContestScoreboardCache(ctx, db, redis, contestID)
}
