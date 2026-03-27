package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func loadAWDContestTeams(ctx context.Context, db *gorm.DB, contestID int64) ([]model.Team, error) {
	var teams []model.Team
	if err := db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func loadAWDServiceScoreRows(ctx context.Context, db *gorm.DB, contestID int64) ([]awdServiceScoreRow, error) {
	var rows []awdServiceScoreRow
	if err := db.WithContext(ctx).
		Table("awd_team_services AS ats").
		Select("ats.team_id AS team_id, ats.defense_score AS defense_score, ats.check_result AS check_result").
		Joins("JOIN awd_rounds AS ar ON ar.id = ats.round_id").
		Where("ar.contest_id = ?", contestID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func loadAWDAttackScoreRows(ctx context.Context, db *gorm.DB, contestID int64) ([]awdAttackScoreRow, error) {
	var rows []awdAttackScoreRow
	if err := db.WithContext(ctx).
		Table("awd_attack_logs AS aal").
		Select("aal.attacker_team_id AS team_id, aal.score_gained AS score_gained, aal.source AS source, aal.created_at AS created_at").
		Joins("JOIN awd_rounds AS ar ON ar.id = aal.round_id").
		Where("ar.contest_id = ? AND aal.score_gained > 0", contestID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
