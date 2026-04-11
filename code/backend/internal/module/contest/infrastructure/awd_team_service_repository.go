package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
)

func (r *AWDRepository) UpsertServiceCheck(
	ctx context.Context,
	roundID, teamID, challengeID int64,
	serviceStatus, checkResult string,
	defenseScore int,
	updatedAt time.Time,
) (*model.AWDTeamService, error) {
	record := &model.AWDTeamService{
		RoundID:       roundID,
		TeamID:        teamID,
		ChallengeID:   challengeID,
		ServiceStatus: serviceStatus,
		CheckResult:   checkResult,
		CheckerType:   "",
		SLAScore:      0,
		DefenseScore:  defenseScore,
	}
	if err := r.dbWithContext(ctx).
		Where("round_id = ? AND team_id = ? AND challenge_id = ?", roundID, teamID, challengeID).
		Assign(map[string]any{
			"service_status": serviceStatus,
			"check_result":   checkResult,
			"defense_score":  defenseScore,
			"updated_at":     updatedAt,
		}).
		FirstOrCreate(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (r *AWDRepository) UpsertTeamServices(ctx context.Context, records []model.AWDTeamService) error {
	if len(records) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "round_id"},
			{Name: "team_id"},
			{Name: "challenge_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"service_status",
			"check_result",
			"checker_type",
			"sla_score",
			"defense_score",
			"updated_at",
		}),
	}).Create(&records).Error
}

func (r *AWDRepository) ListServicesByRound(ctx context.Context, roundID int64) ([]model.AWDTeamService, error) {
	var records []model.AWDTeamService
	err := r.dbWithContext(ctx).
		Where("round_id = ?", roundID).
		Order("team_id ASC, challenge_id ASC").
		Find(&records).Error
	return records, err
}
