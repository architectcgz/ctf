package infrastructure

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type scoreboardTeamStatsRow struct {
	TeamID              int64  `gorm:"column:team_id"`
	SolvedCount         int    `gorm:"column:solved_count"`
	LastSubmissionAtRaw string `gorm:"column:last_submission_at"`
}

func (r *Repository) FindScoreboardTeamStats(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) (map[int64]contestports.ScoreboardTeamStats, error) {
	result := make(map[int64]contestports.ScoreboardTeamStats, len(teamIDs))
	if len(teamIDs) == 0 {
		return result, nil
	}

	var rows []scoreboardTeamStatsRow
	switch contestMode {
	case model.ContestModeAWD:
		if err := r.db.WithContext(ctx).
			Table("awd_attack_logs AS aal").
			Select("aal.attacker_team_id AS team_id, COUNT(*) AS solved_count, MAX(aal.created_at) AS last_submission_at").
			Joins("JOIN awd_rounds AS ar ON ar.id = aal.round_id").
			Where("ar.contest_id = ? AND aal.is_success = ? AND aal.source = ?", contestID, true, model.AWDAttackSourceSubmission).
			Where("aal.attacker_team_id IN ?", teamIDs).
			Group("aal.attacker_team_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
	default:
		if err := r.db.WithContext(ctx).
			Table("submissions").
			Select("team_id AS team_id, COUNT(*) AS solved_count, MAX(submitted_at) AS last_submission_at").
			Where("contest_id = ? AND is_correct = ? AND team_id IS NOT NULL", contestID, true).
			Where("team_id IN ?", teamIDs).
			Group("team_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
	}

	for _, row := range rows {
		result[row.TeamID] = contestports.ScoreboardTeamStats{
			SolvedCount:      row.SolvedCount,
			LastSubmissionAt: parseContestAggregateTime(row.LastSubmissionAtRaw),
		}
	}
	return result, nil
}

func parseContestAggregateTime(raw string) *time.Time {
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
