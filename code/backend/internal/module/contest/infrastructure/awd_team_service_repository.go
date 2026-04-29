package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
)

func (r *AWDRepository) UpsertServiceCheck(
	ctx context.Context,
	roundID, teamID, serviceID, awdChallengeID int64,
	serviceStatus, checkResult string,
	defenseScore int,
	updatedAt time.Time,
) (*model.AWDTeamService, error) {
	record := &model.AWDTeamService{
		RoundID:        roundID,
		TeamID:         teamID,
		ServiceID:      serviceID,
		AWDChallengeID: awdChallengeID,
		ServiceStatus:  serviceStatus,
		CheckResult:    checkResult,
		CheckerType:    "",
		SLAScore:       0,
		DefenseScore:   defenseScore,
	}
	if err := r.dbWithContext(ctx).
		Where("round_id = ? AND team_id = ? AND service_id = ?", roundID, teamID, serviceID).
		Assign(map[string]any{
			"service_id":       serviceID,
			"awd_challenge_id": awdChallengeID,
			"service_status":   serviceStatus,
			"check_result":     checkResult,
			"defense_score":    defenseScore,
			"updated_at":       updatedAt,
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
			{Name: "service_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"awd_challenge_id",
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
		Order("team_id ASC, service_id ASC, awd_challenge_id ASC").
		Find(&records).Error
	return records, err
}
