package infrastructure

import (
	"context"
	"errors"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"gorm.io/gorm"
)

type runtimeProxyTrafficInstanceRow struct {
	ContestID   *int64 `gorm:"column:contest_id"`
	TeamID      *int64 `gorm:"column:team_id"`
	ChallengeID int64  `gorm:"column:challenge_id"`
}

func (r *AWDRepository) RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error {
	instanceScope, err := r.loadRuntimeProxyTrafficInstanceScope(ctx, instanceID)
	if err != nil || instanceScope == nil {
		return err
	}
	if instanceScope.ContestID == nil || instanceScope.TeamID == nil || instanceScope.ChallengeID <= 0 {
		return nil
	}

	round, err := r.FindRunningRound(ctx, *instanceScope.ContestID)
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
		ChallengeID:    instanceScope.ChallengeID,
		Method:         trimToLength(method, 16),
		Path:           trimToLength(requestPath, 1024),
		StatusCode:     statusCode,
		Source:         model.AWDTrafficSourceRuntimeProxy,
	}).Error
}

func (r *AWDRepository) ListTrafficEvents(ctx context.Context, contestID, roundID int64) ([]contestports.AWDTrafficEventRecord, error) {
	rows := make([]contestports.AWDTrafficEventRecord, 0)
	err := r.dbWithContext(ctx).
		Table("awd_traffic_events AS te").
		Select(`
			te.id AS id,
			te.contest_id AS contest_id,
			te.round_id AS round_id,
			te.attacker_team_id AS attacker_team_id,
			COALESCE(att.name, '') AS attacker_team_name,
			te.victim_team_id AS victim_team_id,
			COALESCE(vic.name, '') AS victim_team_name,
			te.challenge_id AS challenge_id,
			COALESCE(ch.title, '') AS challenge_title,
			te.method AS method,
			te.path AS path,
			te.status_code AS status_code,
			te.source AS source,
			te.created_at AS occurred_at
		`).
		Joins("LEFT JOIN teams att ON att.id = te.attacker_team_id").
		Joins("LEFT JOIN teams vic ON vic.id = te.victim_team_id").
		Joins("LEFT JOIN challenges ch ON ch.id = te.challenge_id").
		Where("te.contest_id = ? AND te.round_id = ?", contestID, roundID).
		Order("te.created_at DESC, te.id DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *AWDRepository) loadRuntimeProxyTrafficInstanceScope(ctx context.Context, instanceID int64) (*runtimeProxyTrafficInstanceRow, error) {
	var row runtimeProxyTrafficInstanceRow
	err := r.dbWithContext(ctx).
		Table("instances").
		Select("contest_id, team_id, challenge_id").
		Where("id = ?", instanceID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *AWDRepository) findRuntimeProxyAttackerTeam(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	team, err := r.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return team, nil
}

func trimToLength(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
