package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (r *AWDRepository) ListServiceInstancesByContest(ctx context.Context, contestID int64, challengeIDs []int64) ([]contestports.AWDServiceInstance, error) {
	if len(challengeIDs) == 0 {
		return nil, nil
	}

	var instances []contestports.AWDServiceInstance
	if err := r.dbWithContext(ctx).
		Table("instances AS inst").
		Select("COALESCE(inst.team_id, tm.team_id) AS team_id, inst.challenge_id AS challenge_id, inst.access_url AS access_url").
		Joins("LEFT JOIN team_members AS tm ON tm.user_id = inst.user_id AND tm.contest_id = ?", contestID).
		Where("inst.challenge_id IN ?", challengeIDs).
		Where("inst.status = ?", model.InstanceStatusRunning).
		Where("(inst.contest_id = ? AND inst.team_id IS NOT NULL) OR (inst.team_id IS NULL AND tm.team_id IS NOT NULL)", contestID).
		Order("tm.team_id ASC, inst.challenge_id ASC, inst.id ASC").
		Scan(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

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

func (r *AWDRepository) CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, challengeID int64) (int64, error) {
	var count int64
	err := r.dbWithContext(ctx).
		Model(&model.AWDAttackLog{}).
		Where(
			"round_id = ? AND attacker_team_id = ? AND victim_team_id = ? AND challenge_id = ? AND is_success = ?",
			roundID, attackerTeamID, victimTeamID, challengeID, true,
		).
		Count(&count).Error
	return count, err
}

func (r *AWDRepository) CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error {
	return r.dbWithContext(ctx).Create(logRecord).Error
}

func (r *AWDRepository) ApplyAttackImpactToVictimService(
	ctx context.Context,
	roundID, victimTeamID, challengeID int64,
	scoreGained int,
	updatedAt time.Time,
) error {
	record := &model.AWDTeamService{
		RoundID:        roundID,
		TeamID:         victimTeamID,
		ChallengeID:    challengeID,
		ServiceStatus:  model.AWDServiceStatusCompromised,
		CheckResult:    "{}",
		AttackReceived: 1,
		DefenseScore:   0,
		AttackScore:    scoreGained,
		CreatedAt:      updatedAt,
		UpdatedAt:      updatedAt,
	}
	return r.dbWithContext(ctx).
		Where("round_id = ? AND team_id = ? AND challenge_id = ?", roundID, victimTeamID, challengeID).
		Assign(map[string]any{
			"service_status":  model.AWDServiceStatusCompromised,
			"attack_received": gorm.Expr("attack_received + ?", 1),
			"attack_score":    gorm.Expr("attack_score + ?", scoreGained),
			"defense_score":   0,
			"updated_at":      updatedAt,
		}).
		FirstOrCreate(record).Error
}

func (r *AWDRepository) ListAttackLogsByRound(ctx context.Context, roundID int64) ([]model.AWDAttackLog, error) {
	var logs []model.AWDAttackLog
	err := r.dbWithContext(ctx).
		Where("round_id = ?", roundID).
		Order("created_at ASC, id ASC").
		Find(&logs).Error
	return logs, err
}
