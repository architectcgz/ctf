package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *AWDRepository) FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*contestentity.AWDDefenseWorkspace, error) {
	var workspace contestentity.AWDDefenseWorkspace
	err := r.dbWithContext(ctx).
		Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
		First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workspace, nil
}

func (r *AWDRepository) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *contestentity.AWDDefenseWorkspace) error {
	if workspace == nil {
		return nil
	}

	if workspace.WorkspaceRevision <= 0 {
		workspace.WorkspaceRevision = 1
	}
	if strings.TrimSpace(workspace.Status) == "" {
		workspace.Status = contestentity.AWDDefenseWorkspaceStatusPending
	}

	now := time.Now().UTC()
	if err := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "contest_id"},
			{Name: "team_id"},
			{Name: "service_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"instance_id":        workspace.InstanceID,
			"workspace_revision": workspace.WorkspaceRevision,
			"status":             workspace.Status,
			"container_id":       workspace.ContainerID,
			"seed_signature":     workspace.SeedSignature,
			"updated_at":         now,
		}),
	}).Create(workspace).Error; err != nil {
		return err
	}

	stored, err := r.FindAWDDefenseWorkspace(ctx, workspace.ContestID, workspace.TeamID, workspace.ServiceID)
	if err != nil {
		return err
	}
	if stored != nil {
		*workspace = *stored
	}
	return nil
}

func (r *AWDRepository) BumpAWDDefenseWorkspaceRevision(ctx context.Context, contestID, teamID, serviceID, instanceID int64, seedSignature string) error {
	now := time.Now().UTC()
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var workspace contestentity.AWDDefenseWorkspace
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
			First(&workspace).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return tx.Create(&contestentity.AWDDefenseWorkspace{
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				InstanceID:        instanceID,
				WorkspaceRevision: 1,
				Status:            contestentity.AWDDefenseWorkspaceStatusProvisioning,
				SeedSignature:     seedSignature,
				CreatedAt:         now,
				UpdatedAt:         now,
			}).Error
		}

		return tx.Model(&contestentity.AWDDefenseWorkspace{}).
			Where("id = ?", workspace.ID).
			Updates(map[string]any{
				"instance_id":        instanceID,
				"workspace_revision": workspace.WorkspaceRevision + 1,
				"status":             contestentity.AWDDefenseWorkspaceStatusProvisioning,
				"container_id":       "",
				"seed_signature":     seedSignature,
				"updated_at":         now,
			}).Error
	})
}

func (r *AWDRepository) FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*contestentity.AWDDefenseWorkspace, error) {
	if instanceID <= 0 {
		return nil, nil
	}

	var workspace contestentity.AWDDefenseWorkspace
	err := r.dbWithContext(ctx).
		Where("instance_id = ?", instanceID).
		Where("status = ?", contestentity.AWDDefenseWorkspaceStatusRunning).
		Where("container_id <> ''").
		First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist") {
			return nil, nil
		}
		return nil, err
	}
	return &workspace, nil
}

func (r *AWDRepository) CreateAWDServiceOperation(ctx context.Context, operation *contestentity.AWDServiceOperation) error {
	return r.dbWithContext(ctx).Create(operation).Error
}

func (r *AWDRepository) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&contestentity.AWDServiceOperation{}).
		Where("instance_id = ? AND status IN ?", instanceID, []string{
			contestentity.AWDServiceOperationStatusRequested,
			contestentity.AWDServiceOperationStatusProvisioning,
			contestentity.AWDServiceOperationStatusRecovering,
		}).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now().UTC(),
		}).Error
}

func (r *AWDRepository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if operationID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&contestentity.AWDServiceOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now().UTC(),
		}).Error
}
