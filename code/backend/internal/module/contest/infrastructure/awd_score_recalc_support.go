package infrastructure

import (
	"time"

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
	SLAScore     int    `gorm:"column:sla_score"`
	DefenseScore int    `gorm:"column:defense_score"`
	CheckResult  string `gorm:"column:check_result"`
}

type awdServiceScoreTotal struct {
	SLAScore     int
	DefenseScore int
}

func accumulateAWDServiceScores(rows []awdServiceScoreRow) map[int64]awdServiceScoreTotal {
	scoreMap := make(map[int64]awdServiceScoreTotal, len(rows))
	for _, row := range rows {
		if !shouldCountAWDDefenseScoreForOfficialTotals(row.CheckResult) {
			continue
		}
		current := scoreMap[row.TeamID]
		current.SLAScore += row.SLAScore
		current.DefenseScore += row.DefenseScore
		scoreMap[row.TeamID] = current
	}
	return scoreMap
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
