package infrastructure

import (
	"context"

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
		Select("COALESCE(cas.id, 0) AS service_id, COALESCE(inst.team_id, tm.team_id) AS team_id, inst.challenge_id AS challenge_id, inst.access_url AS access_url").
		Joins("LEFT JOIN team_members AS tm ON tm.user_id = inst.user_id AND tm.contest_id = ?", contestID).
		Joins("LEFT JOIN contest_awd_services AS cas ON cas.contest_id = ? AND cas.challenge_id = inst.challenge_id AND cas.deleted_at IS NULL", contestID).
		Where("inst.challenge_id IN ?", challengeIDs).
		Where("inst.status = ?", model.InstanceStatusRunning).
		Where("(inst.contest_id = ? AND inst.team_id IS NOT NULL) OR (inst.team_id IS NULL AND tm.team_id IS NOT NULL)", contestID).
		Order("tm.team_id ASC, inst.challenge_id ASC, inst.id ASC").
		Scan(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}
