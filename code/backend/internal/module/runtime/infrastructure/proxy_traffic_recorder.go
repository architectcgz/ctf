package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func NewProxyTrafficEventRecorder(db *gorm.DB) *Repository {
	return NewRepository(db)
}

type runtimeProxyTrafficInstanceRow struct {
	ContestID   *int64 `gorm:"column:contest_id"`
	TeamID      *int64 `gorm:"column:team_id"`
	ServiceID   *int64 `gorm:"column:service_id"`
	ChallengeID int64  `gorm:"column:challenge_id"`
}

func (r *Repository) RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error {
	instanceScope, err := r.loadRuntimeProxyTrafficInstanceScope(ctx, instanceID)
	if err != nil || instanceScope == nil {
		return err
	}
	if instanceScope.ContestID == nil || instanceScope.TeamID == nil || instanceScope.ServiceID == nil || *instanceScope.ServiceID <= 0 || instanceScope.ChallengeID <= 0 {
		return nil
	}

	round, err := r.findRunningAWDRound(ctx, *instanceScope.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	attackerTeam, err := r.findRuntimeProxyAttackerTeam(ctx, *instanceScope.ContestID, userID)
	if err != nil {
		return err
	}
	if attackerTeam == nil {
		return nil
	}

	return r.dbWithContext(ctx).Create(&model.AWDTrafficEvent{
		ContestID:      *instanceScope.ContestID,
		RoundID:        round.ID,
		AttackerTeamID: attackerTeam.ID,
		VictimTeamID:   *instanceScope.TeamID,
		ServiceID:      *instanceScope.ServiceID,
		ChallengeID:    instanceScope.ChallengeID,
		Method:         trimProxyTrafficField(method, 16),
		Path:           trimProxyTrafficField(requestPath, 1024),
		StatusCode:     statusCode,
		Source:         model.AWDTrafficSourceRuntimeProxy,
	}).Error
}

func (r *Repository) RecordAWDProxyTrafficEvent(ctx context.Context, event model.AWDProxyTrafficEventInput) error {
	if event.ContestID <= 0 || event.AttackerTeamID <= 0 || event.VictimTeamID <= 0 || event.ServiceID <= 0 || event.ChallengeID <= 0 {
		return nil
	}

	round, err := r.findRunningAWDRound(ctx, event.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return r.dbWithContext(ctx).Create(&model.AWDTrafficEvent{
		ContestID:      event.ContestID,
		RoundID:        round.ID,
		AttackerTeamID: event.AttackerTeamID,
		VictimTeamID:   event.VictimTeamID,
		ServiceID:      event.ServiceID,
		ChallengeID:    event.ChallengeID,
		Method:         trimProxyTrafficField(event.Method, 16),
		Path:           trimProxyTrafficField(event.Path, 1024),
		StatusCode:     event.StatusCode,
		Source:         model.AWDTrafficSourceRuntimeProxy,
	}).Error
}

func (r *Repository) loadRuntimeProxyTrafficInstanceScope(ctx context.Context, instanceID int64) (*runtimeProxyTrafficInstanceRow, error) {
	var row runtimeProxyTrafficInstanceRow
	err := r.dbWithContext(ctx).
		Table("instances AS inst").
		Select("inst.contest_id, inst.team_id, inst.service_id, cas.challenge_id AS challenge_id").
		Joins("LEFT JOIN contest_awd_services AS cas ON cas.id = inst.service_id AND cas.deleted_at IS NULL").
		Where("inst.id = ?", instanceID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *Repository) findRunningAWDRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	var round model.AWDRound
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND status = ?", contestID, model.AWDRoundStatusRunning).
		Order("round_number DESC, id DESC").
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *Repository) findRuntimeProxyAttackerTeam(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	var team model.Team
	if err := r.dbWithContext(ctx).
		Table("teams AS t").
		Select("t.*").
		Joins("JOIN team_members AS tm ON tm.team_id = t.id").
		Where("t.contest_id = ? AND tm.user_id = ?", contestID, userID).
		First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func trimProxyTrafficField(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
