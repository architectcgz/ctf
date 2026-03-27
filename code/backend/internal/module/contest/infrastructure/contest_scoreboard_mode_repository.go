package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

func (r *Repository) findScoreboardTeamStatsRows(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) ([]scoreboardTeamStatsRow, error) {
	switch contestMode {
	case model.ContestModeAWD:
		return r.findAWDScoreboardTeamStatsRows(ctx, contestID, teamIDs)
	default:
		return r.findStandardScoreboardTeamStatsRows(ctx, contestID, teamIDs)
	}
}

func (r *Repository) findAWDScoreboardTeamStatsRows(ctx context.Context, contestID int64, teamIDs []int64) ([]scoreboardTeamStatsRow, error) {
	var rows []scoreboardTeamStatsRow
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
	return rows, nil
}

func (r *Repository) findStandardScoreboardTeamStatsRows(ctx context.Context, contestID int64, teamIDs []int64) ([]scoreboardTeamStatsRow, error) {
	var rows []scoreboardTeamStatsRow
	if err := r.db.WithContext(ctx).
		Table("submissions").
		Select("team_id AS team_id, COUNT(*) AS solved_count, MAX(submitted_at) AS last_submission_at").
		Where("contest_id = ? AND is_correct = ? AND team_id IS NOT NULL", contestID, true).
		Where("team_id IN ?", teamIDs).
		Group("team_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
