package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

type RuntimeStateRepository struct {
	db *gorm.DB
}

func NewRuntimeStateRepository(db *gorm.DB) *RuntimeStateRepository {
	return &RuntimeStateRepository{db: db}
}

func (r *RuntimeStateRepository) WithDB(db *gorm.DB) *RuntimeStateRepository {
	return &RuntimeStateRepository{db: db}
}

func (r *RuntimeStateRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *RuntimeStateRepository) FindByID(ctx context.Context, id int64) (*runtimecontracts.RuntimeManagedInstance, error) {
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

func (r *RuntimeStateRepository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
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

func (r *RuntimeStateRepository) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
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

func (r *RuntimeStateRepository) ListInstancesNeedingACLHandleMigration(ctx context.Context) ([]runtimecontracts.RuntimeManagedInstance, error) {
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

func (r *RuntimeStateRepository) UpdateInstanceRuntimeDetails(ctx context.Context, instanceID int64, runtimeDetails string) error {
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
