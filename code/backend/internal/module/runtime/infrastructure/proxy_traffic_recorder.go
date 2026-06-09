package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
)

type ProxyTrafficEventRecorder struct {
	db *gorm.DB
}

func NewProxyTrafficEventRecorder(db *gorm.DB) *ProxyTrafficEventRecorder {
	return &ProxyTrafficEventRecorder{db: db}
}

func (r *ProxyTrafficEventRecorder) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

type runtimeProxyTrafficInstanceRow struct {
	ContestID      *int64 `gorm:"column:contest_id"`
	TeamID         *int64 `gorm:"column:team_id"`
	ServiceID      *int64 `gorm:"column:service_id"`
	AWDChallengeID int64  `gorm:"column:awd_challenge_id"`
}

func (r *ProxyTrafficEventRecorder) RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error {
	instanceScope, err := r.loadRuntimeProxyTrafficInstanceScope(ctx, instanceID)
	if err != nil || instanceScope == nil {
		return err
	}
	if instanceScope.ContestID == nil || instanceScope.TeamID == nil || instanceScope.ServiceID == nil || *instanceScope.ServiceID <= 0 || instanceScope.AWDChallengeID <= 0 {
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

	return r.dbWithContext(ctx).Create(&contestcontracts.AWDTrafficEvent{
		ContestID:      *instanceScope.ContestID,
		RoundID:        round.ID,
		AttackerTeamID: attackerTeam.ID,
		VictimTeamID:   *instanceScope.TeamID,
		ServiceID:      *instanceScope.ServiceID,
		AWDChallengeID: instanceScope.AWDChallengeID,
		Method:         trimProxyTrafficField(method, 16),
		Path:           trimProxyTrafficField(requestPath, 1024),
		StatusCode:     statusCode,
		Source:         contestcontracts.AWDTrafficSourceRuntimeProxy,
	}).Error
}

func (r *ProxyTrafficEventRecorder) RecordAWDProxyTrafficEvent(ctx context.Context, event contestcontracts.AWDProxyTrafficEventInput) error {
	if event.ContestID <= 0 || event.AttackerTeamID <= 0 || event.VictimTeamID <= 0 || event.ServiceID <= 0 || event.AWDChallengeID <= 0 {
		return nil
	}

	round, err := r.findRunningAWDRound(ctx, event.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return r.dbWithContext(ctx).Create(&contestcontracts.AWDTrafficEvent{
		ContestID:      event.ContestID,
		RoundID:        round.ID,
		AttackerTeamID: event.AttackerTeamID,
		VictimTeamID:   event.VictimTeamID,
		ServiceID:      event.ServiceID,
		AWDChallengeID: event.AWDChallengeID,
		Method:         trimProxyTrafficField(event.Method, 16),
		Path:           trimProxyTrafficField(event.Path, 1024),
		StatusCode:     event.StatusCode,
		Source:         contestcontracts.AWDTrafficSourceRuntimeProxy,
	}).Error
}

func (r *ProxyTrafficEventRecorder) loadRuntimeProxyTrafficInstanceScope(ctx context.Context, instanceID int64) (*runtimeProxyTrafficInstanceRow, error) {
	var row runtimeProxyTrafficInstanceRow
	err := r.dbWithContext(ctx).
		Table("instances AS inst").
		Select("inst.contest_id, inst.team_id, inst.service_id, cas.awd_challenge_id AS awd_challenge_id").
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

func (r *ProxyTrafficEventRecorder) findRunningAWDRound(ctx context.Context, contestID int64) (*contestcontracts.AWDRound, error) {
	var round contestcontracts.AWDRound
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND status = ?", contestID, contestcontracts.AWDRoundStatusRunning).
		Order("round_number DESC, id DESC").
		First(&round).Error; err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *ProxyTrafficEventRecorder) findRuntimeProxyAttackerTeam(ctx context.Context, contestID, userID int64) (*contestcontracts.Team, error) {
	var team contestcontracts.Team
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
