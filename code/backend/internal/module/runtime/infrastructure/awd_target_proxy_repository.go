package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func (r *Repository) FindAWDTargetProxyScope(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*runtimeports.AWDTargetProxyScope, error) {
	if userID <= 0 || contestID <= 0 || serviceID <= 0 || victimTeamID <= 0 {
		return nil, nil
	}

	var scope runtimeports.AWDTargetProxyScope
	err := r.dbWithContext(ctx).
		Table("contests AS co").
		Select(`
			inst.id AS instance_id,
			inst.access_url AS access_url,
			inst.share_scope AS share_scope,
			co.id AS contest_id,
			tm.team_id AS attacker_team_id,
			victim.id AS victim_team_id,
			cas.id AS service_id,
			cas.awd_challenge_id AS awd_challenge_id
		`).
		Joins("JOIN team_members AS tm ON tm.contest_id = co.id AND tm.user_id = ?", userID).
		Joins("JOIN teams AS victim ON victim.contest_id = co.id AND victim.id = ? AND victim.deleted_at IS NULL", victimTeamID).
		Joins("JOIN contest_awd_services AS cas ON cas.contest_id = co.id AND cas.id = ? AND cas.is_visible = ? AND cas.deleted_at IS NULL", serviceID, true).
		Joins("JOIN instances AS inst ON inst.contest_id = co.id AND inst.team_id = victim.id AND inst.service_id = cas.id").
		Joins("JOIN awd_rounds AS round ON round.contest_id = co.id AND round.status = ?", model.AWDRoundStatusRunning).
		Where("co.id = ? AND co.mode = ? AND co.status IN ? AND co.deleted_at IS NULL", contestID, model.ContestModeAWD, []string{model.ContestStatusRunning, model.ContestStatusFrozen}).
		Where("tm.team_id <> victim.id").
		Where("inst.status = ? AND inst.access_url <> ''", model.InstanceStatusRunning).
		Order("inst.created_at DESC, inst.id DESC").
		Limit(1).
		Scan(&scope).Error
	if err != nil {
		return nil, err
	}
	if scope.InstanceID <= 0 {
		return nil, nil
	}
	return &scope, nil
}

func (r *Repository) FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*runtimeports.AWDDefenseSSHScope, error) {
	if userID <= 0 || contestID <= 0 || serviceID <= 0 {
		return nil, nil
	}

	var scope runtimeports.AWDDefenseSSHScope
	err := r.dbWithContext(ctx).
		Table("contests AS co").
		Select(`
			inst.id AS instance_id,
			inst.container_id AS container_id,
			inst.share_scope AS share_scope,
			co.id AS contest_id,
			tm.team_id AS team_id,
			cas.id AS service_id,
			cas.awd_challenge_id AS awd_challenge_id
		`).
		Joins("JOIN team_members AS tm ON tm.contest_id = co.id AND tm.user_id = ?", userID).
		Joins("JOIN contest_awd_services AS cas ON cas.contest_id = co.id AND cas.id = ? AND cas.is_visible = ? AND cas.deleted_at IS NULL", serviceID, true).
		Joins("JOIN instances AS inst ON inst.contest_id = co.id AND inst.team_id = tm.team_id AND inst.service_id = cas.id").
		Joins("JOIN awd_rounds AS round ON round.contest_id = co.id AND round.status = ?", model.AWDRoundStatusRunning).
		Where("co.id = ? AND co.mode = ? AND co.status IN ? AND co.deleted_at IS NULL", contestID, model.ContestModeAWD, []string{model.ContestStatusRunning, model.ContestStatusFrozen}).
		Where("inst.status = ? AND inst.container_id <> ''", model.InstanceStatusRunning).
		Order("inst.created_at DESC, inst.id DESC").
		Limit(1).
		Scan(&scope).Error
	if err != nil {
		return nil, err
	}
	if scope.InstanceID <= 0 {
		return nil, nil
	}
	return &scope, nil
}
