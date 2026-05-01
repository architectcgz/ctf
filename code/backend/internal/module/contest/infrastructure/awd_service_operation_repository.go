package infrastructure

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

func (r *AWDRepository) ListLatestServiceOperationsByContest(ctx context.Context, contestID int64) ([]model.AWDServiceOperation, error) {
	var operations []model.AWDServiceOperation
	if contestID <= 0 {
		return operations, nil
	}
	if err := r.dbWithContext(ctx).
		Table("awd_service_operations AS op").
		Joins(`JOIN (
			SELECT contest_id, team_id, service_id, MAX(id) AS id
			FROM awd_service_operations
			WHERE contest_id = ?
			GROUP BY contest_id, team_id, service_id
		) latest ON latest.id = op.id`, contestID).
		Order("op.team_id ASC, op.service_id ASC").
		Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *AWDRepository) HasSystemRecoveryOperationAt(ctx context.Context, contestID, teamID, serviceID int64, checkedAt time.Time) (bool, error) {
	if contestID <= 0 || teamID <= 0 || serviceID <= 0 || checkedAt.IsZero() {
		return false, nil
	}
	var count int64
	err := r.dbWithContext(ctx).
		Model(&model.AWDServiceOperation{}).
		Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
		Where("requested_by = ? AND operation_type IN ?", model.AWDServiceOperationRequestedBySystem, []string{
			model.AWDServiceOperationTypeRecover,
			model.AWDServiceOperationTypeRecreate,
		}).
		Where("sla_billable = ?", false).
		Where("(status IN ? OR (started_at <= ? AND (finished_at IS NULL OR finished_at >= ?)))", []string{
			model.AWDServiceOperationStatusRequested,
			model.AWDServiceOperationStatusProvisioning,
			model.AWDServiceOperationStatusRecovering,
		}, checkedAt, checkedAt).
		Count(&count).Error
	return count > 0, err
}
