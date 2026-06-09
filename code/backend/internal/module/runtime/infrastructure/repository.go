package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*runtimecontracts.RuntimeManagedInstance, error) {
	var instance runtimecontracts.RuntimeManagedInstance
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*runtimeentity.AWDDefenseWorkspace, error) {
	var workspace runtimeentity.AWDDefenseWorkspace
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

func (r *Repository) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimeentity.AWDDefenseWorkspace) error {
	if workspace == nil {
		return nil
	}

	if workspace.WorkspaceRevision <= 0 {
		workspace.WorkspaceRevision = 1
	}
	if strings.TrimSpace(workspace.Status) == "" {
		workspace.Status = runtimeentity.AWDDefenseWorkspaceStatusPending
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

func (r *Repository) BumpAWDDefenseWorkspaceRevision(ctx context.Context, contestID, teamID, serviceID, instanceID int64, seedSignature string) error {
	now := time.Now().UTC()
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var workspace runtimeentity.AWDDefenseWorkspace
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
			First(&workspace).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return tx.Create(&runtimeentity.AWDDefenseWorkspace{
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				InstanceID:        instanceID,
				WorkspaceRevision: 1,
				Status:            runtimeentity.AWDDefenseWorkspaceStatusProvisioning,
				SeedSignature:     seedSignature,
				CreatedAt:         now,
				UpdatedAt:         now,
			}).Error
		}

		return tx.Model(&runtimeentity.AWDDefenseWorkspace{}).
			Where("id = ?", workspace.ID).
			Updates(map[string]any{
				"instance_id":        instanceID,
				"workspace_revision": workspace.WorkspaceRevision + 1,
				"status":             runtimeentity.AWDDefenseWorkspaceStatusProvisioning,
				"container_id":       "",
				"seed_signature":     seedSignature,
				"updated_at":         now,
			}).Error
	})
}

func (r *Repository) FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*runtimeentity.AWDDefenseWorkspace, error) {
	if instanceID <= 0 {
		return nil, nil
	}

	var workspace runtimeentity.AWDDefenseWorkspace
	err := r.dbWithContext(ctx).
		Where("instance_id = ?", instanceID).
		Where("status = ?", runtimeentity.AWDDefenseWorkspaceStatusRunning).
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

func (r *Repository) CreateAWDServiceOperation(ctx context.Context, operation *runtimeentity.AWDServiceOperation) error {
	return r.dbWithContext(ctx).Create(operation).Error
}

func (r *Repository) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&runtimeentity.AWDServiceOperation{}).
		Where("instance_id = ? AND status IN ?", instanceID, []string{
			runtimeentity.AWDServiceOperationStatusRequested,
			runtimeentity.AWDServiceOperationStatusProvisioning,
			runtimeentity.AWDServiceOperationStatusRecovering,
		}).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now().UTC(),
		}).Error
}

func (r *Repository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if operationID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&runtimeentity.AWDServiceOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now().UTC(),
		}).Error
}

func (r *Repository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	var items []struct {
		ContainerID    string
		RuntimeDetails string
	}
	if err := r.dbWithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("status IN ?", []string{
			runtimecontracts.RuntimeManagedInstanceStatusCreating,
			runtimecontracts.RuntimeManagedInstanceStatusRunning,
			runtimecontracts.RuntimeManagedInstanceStatusStopping,
		}).
		Select("container_id, runtime_details").
		Scan(&items).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		ids := []string{item.ContainerID}
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(item.RuntimeDetails)
		if err == nil {
			for _, container := range details.Containers {
				ids = append(ids, container.ContainerID)
			}
		}
		for _, containerID := range ids {
			if containerID == "" {
				continue
			}
			if _, exists := seen[containerID]; exists {
				continue
			}
			seen[containerID] = struct{}{}
			result = append(result, containerID)
		}
	}

	var workspaceItems []struct {
		ContainerID string
	}
	if err := r.dbWithContext(ctx).
		Table("awd_defense_workspaces AS ws").
		Joins("JOIN instances AS inst ON inst.id = ws.instance_id").
		Where("inst.status IN ?", []string{
			runtimecontracts.RuntimeManagedInstanceStatusCreating,
			runtimecontracts.RuntimeManagedInstanceStatusRunning,
			runtimecontracts.RuntimeManagedInstanceStatusStopping,
		}).
		Where("ws.status = ? AND ws.container_id <> ''", runtimeentity.AWDDefenseWorkspaceStatusRunning).
		Select("ws.container_id").
		Scan(&workspaceItems).Error; err != nil {
		lowerErr := strings.ToLower(err.Error())
		if !strings.Contains(lowerErr, "no such table") && !strings.Contains(lowerErr, "does not exist") {
			return nil, err
		}
		return result, nil
	}
	for _, item := range workspaceItems {
		containerID := strings.TrimSpace(item.ContainerID)
		if containerID == "" {
			continue
		}
		if _, exists := seen[containerID]; exists {
			continue
		}
		seen[containerID] = struct{}{}
		result = append(result, containerID)
	}
	return result, nil
}

func (r *Repository) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, nil
	}

	type instanceContainerLookupRow struct {
		NodeID         *int64 `gorm:"column:node_id"`
		ContainerID    string `gorm:"column:container_id"`
		RuntimeDetails string `gorm:"column:runtime_details"`
	}

	rows := make([]instanceContainerLookupRow, 0)
	likePattern := "%" + containerID + "%"
	if err := r.dbWithContext(ctx).
		Model(&runtimecontracts.RuntimeManagedInstance{}).
		Select("node_id, container_id, runtime_details").
		Where("container_id = ? OR runtime_details LIKE ?", containerID, likePattern).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if row.ContainerID == containerID {
			return row.NodeID, nil
		}
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(row.RuntimeDetails)
		if err != nil {
			continue
		}
		for _, item := range details.Containers {
			if strings.TrimSpace(item.ContainerID) == containerID {
				return row.NodeID, nil
			}
		}
	}

	type workspaceContainerLookupRow struct {
		NodeID *int64 `gorm:"column:node_id"`
	}
	var workspace workspaceContainerLookupRow
	if err := r.dbWithContext(ctx).
		Table("awd_defense_workspaces AS ws").
		Joins("JOIN instances AS inst ON inst.id = ws.instance_id").
		Where("ws.container_id = ?", containerID).
		Select("inst.node_id").
		Take(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist") {
			return nil, nil
		}
		return nil, err
	}
	return workspace.NodeID, nil
}

func (r *Repository) ListInstancesNeedingACLHandleMigration(ctx context.Context) ([]runtimecontracts.RuntimeManagedInstance, error) {
	type instanceACLMigrationRow struct {
		ID             int64  `gorm:"column:id"`
		NodeID         *int64 `gorm:"column:node_id"`
		RuntimeDetails string `gorm:"column:runtime_details"`
	}

	rows := make([]instanceACLMigrationRow, 0)
	if err := r.dbWithContext(ctx).
		Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("destroyed_at IS NULL").
		Where("runtime_details <> ''").
		Select("id, node_id, runtime_details").
		Scan(&rows).Error; err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist") {
			return nil, nil
		}
		return nil, err
	}

	result := make([]runtimecontracts.RuntimeManagedInstance, 0, len(rows))
	for _, row := range rows {
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(row.RuntimeDetails)
		if err != nil || details.ACL != nil || len(details.ACLRules) == 0 {
			continue
		}
		result = append(result, runtimecontracts.RuntimeManagedInstance{
			ID:             row.ID,
			NodeID:         row.NodeID,
			RuntimeDetails: row.RuntimeDetails,
		})
	}
	return result, nil
}

func (r *Repository) UpdateInstanceRuntimeDetails(ctx context.Context, instanceID int64, runtimeDetails string) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"runtime_details": runtimeDetails,
			"updated_at":      time.Now().UTC(),
		}).Error
}
