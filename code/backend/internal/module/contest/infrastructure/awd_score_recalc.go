package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

type awdDefenseScoreRow struct {
	TeamID       int64 `gorm:"column:team_id"`
	DefenseScore int   `gorm:"column:defense_score"`
}

type awdAttackScoreRow struct {
	TeamID      int64     `gorm:"column:team_id"`
	ScoreGained int       `gorm:"column:score_gained"`
	Source      string    `gorm:"column:source"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

type awdServiceScoreRow struct {
	TeamID       int64  `gorm:"column:team_id"`
	DefenseScore int    `gorm:"column:defense_score"`
	CheckResult  string `gorm:"column:check_result"`
}

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

func accumulateAWDDefenseScores(rows []awdServiceScoreRow) map[int64]int {
	defenseMap := make(map[int64]int, len(rows))
	for _, row := range rows {
		if !shouldCountAWDDefenseScoreForOfficialTotals(row.CheckResult) {
			continue
		}
		defenseMap[row.TeamID] += row.DefenseScore
	}
	return defenseMap
}

func accumulateAWDAttackScores(rows []awdAttackScoreRow) map[int64]awdAttackScoreRow {
	attackMap := make(map[int64]awdAttackScoreRow, len(rows))
	for _, row := range rows {
		if !shouldCountAWDAttackForOfficialTotals(row.Source) {
			continue
		}
		current := attackMap[row.TeamID]
		current.TeamID = row.TeamID
		current.ScoreGained += row.ScoreGained
		if current.CreatedAt.IsZero() || row.CreatedAt.After(current.CreatedAt) {
			current.CreatedAt = row.CreatedAt
		}
		attackMap[row.TeamID] = current
	}
	return attackMap
}

func shouldCountAWDDefenseScoreForOfficialTotals(checkResult string) bool {
	switch normalizeAWDCheckSourceValue(parseAWDCheckResultValue(checkResult)["check_source"]) {
	case "scheduler", "manual_current_round", "manual_selected_round", "manual_service_check":
		return true
	default:
		return false
	}
}

func shouldCountAWDAttackForOfficialTotals(source string) bool {
	return normalizeAWDAttackSourceValue(source) == model.AWDAttackSourceSubmission
}
