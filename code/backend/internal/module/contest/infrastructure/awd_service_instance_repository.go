package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (r *AWDRepository) ListServiceInstancesByContest(ctx context.Context, contestID int64, serviceIDs []int64) ([]contestports.AWDServiceInstance, error) {
	if len(serviceIDs) == 0 {
		return nil, nil
	}

	var instances []contestports.AWDServiceInstance
	if err := r.dbWithContext(ctx).
		Table("instances AS inst").
		Select("inst.id AS instance_id, cas.id AS service_id, COALESCE(inst.team_id, tm.team_id) AS team_id, cas.awd_challenge_id AS awd_challenge_id, inst.host_port AS host_port, inst.container_id AS container_id, inst.network_id AS network_id, inst.status AS status, inst.access_url AS access_url, inst.runtime_details AS runtime_details").
		Joins("LEFT JOIN team_members AS tm ON tm.user_id = inst.user_id AND tm.contest_id = ?", contestID).
		Joins("JOIN contest_awd_services AS cas ON cas.contest_id = ? AND cas.id = inst.service_id AND cas.deleted_at IS NULL", contestID).
		Where("cas.id IN ?", serviceIDs).
		Where("inst.status IN ?", []string{
			model.InstanceStatusPending,
			model.InstanceStatusCreating,
			model.InstanceStatusRunning,
			model.InstanceStatusFailed,
		}).
		Where("(inst.contest_id = ? AND inst.team_id IS NOT NULL) OR (inst.team_id IS NULL AND tm.team_id IS NOT NULL)", contestID).
		Order("COALESCE(inst.team_id, tm.team_id) ASC, cas.\"order\" ASC, cas.id ASC, inst.created_at DESC, inst.id DESC").
		Scan(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}
