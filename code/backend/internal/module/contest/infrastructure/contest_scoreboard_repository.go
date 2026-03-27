package infrastructure

import (
	"context"

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

	rows, err := r.findScoreboardTeamStatsRows(ctx, contestID, contestMode, teamIDs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.TeamID] = contestports.ScoreboardTeamStats{
			SolvedCount:      row.SolvedCount,
			LastSubmissionAt: parseContestAggregateTime(row.LastSubmissionAtRaw),
		}
	}
	return result, nil
}
