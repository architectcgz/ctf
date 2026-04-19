package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func (r *AWDRepository) CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, serviceID int64) (int64, error) {
	var count int64
	err := r.dbWithContext(ctx).
		Model(&model.AWDAttackLog{}).
		Where(
			"round_id = ? AND attacker_team_id = ? AND victim_team_id = ? AND service_id = ? AND is_success = ?",
			roundID, attackerTeamID, victimTeamID, serviceID, true,
		).
		Count(&count).Error
	return count, err
}

func (r *AWDRepository) CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error {
	return r.dbWithContext(ctx).Create(logRecord).Error
}

func (r *AWDRepository) ApplyAttackImpactToVictimService(
	ctx context.Context,
	roundID, victimTeamID, serviceID, challengeID int64,
	scoreGained int,
	updatedAt time.Time,
) error {
	record := &model.AWDTeamService{
		RoundID:        roundID,
		TeamID:         victimTeamID,
		ServiceID:      serviceID,
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
		Where("round_id = ? AND team_id = ? AND service_id = ?", roundID, victimTeamID, serviceID).
		Assign(map[string]any{
			"service_id":      serviceID,
			"challenge_id":    challengeID,
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
