package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
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

	defenseMap := make(map[int64]int, len(serviceRows))
	for _, row := range serviceRows {
		if !shouldCountAWDDefenseScoreForOfficialTotals(row.CheckResult) {
			continue
		}
		defenseMap[row.TeamID] += row.DefenseScore
	}

	attackMap := make(map[int64]awdAttackScoreRow, len(attackRows))
	for _, row := range attackRows {
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

func RebuildContestScoreboardCache(ctx context.Context, db *gorm.DB, redis *redislib.Client, contestID int64) error {
	if db == nil || redis == nil || contestID <= 0 {
		return nil
	}

	var teams []model.Team
	if err := db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error; err != nil {
		return err
	}

	key := rediskeys.RankContestTeamKey(contestID)
	pipe := redis.TxPipeline()
	pipe.Del(ctx, key)

	entries := make([]redislib.Z, 0, len(teams))
	for _, team := range teams {
		if team.TotalScore <= 0 {
			continue
		}
		entries = append(entries, redislib.Z{
			Score:  float64(team.TotalScore),
			Member: contestdomain.TeamIDToMember(team.ID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, key, entries...)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func SyncAWDContestScores(ctx context.Context, db *gorm.DB, redis *redislib.Client, contestID int64) error {
	if err := RecalculateAWDContestTeamScores(ctx, db, contestID); err != nil {
		return err
	}
	return RebuildContestScoreboardCache(ctx, db, redis, contestID)
}

func normalizeAWDCheckSourceValue(value any) string {
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	switch strings.TrimSpace(raw) {
	case "scheduler":
		return "scheduler"
	case "manual_current_round":
		return "manual_current_round"
	case "manual_selected_round":
		return "manual_selected_round"
	case "manual_service_check":
		return "manual_service_check"
	default:
		return ""
	}
}

func parseAWDCheckResultValue(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return map[string]any{}
	}
	return result
}

func normalizeAWDAttackSourceValue(value string) string {
	switch strings.TrimSpace(value) {
	case model.AWDAttackSourceManual:
		return model.AWDAttackSourceManual
	case model.AWDAttackSourceSubmission:
		return model.AWDAttackSourceSubmission
	default:
		return model.AWDAttackSourceLegacy
	}
}

func parseAWDScoreSyncTime(raw string) *time.Time {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return &parsed
		}
	}
	return nil
}
