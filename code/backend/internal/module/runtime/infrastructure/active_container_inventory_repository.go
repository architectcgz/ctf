package infrastructure

import (
	"context"
	"strings"

	"gorm.io/gorm"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimestate "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

type ActiveContainerInventoryRepository struct {
	db *gorm.DB
}

func NewActiveContainerInventoryRepository(db *gorm.DB) *ActiveContainerInventoryRepository {
	return &ActiveContainerInventoryRepository{db: db}
}

func (r *ActiveContainerInventoryRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ActiveContainerInventoryRepository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	var items []struct {
		ContainerID    string
		RuntimeDetails string
	}
	if err := r.dbWithContext(ctx).Model(&runtimestate.RuntimeManagedInstance{}).
		Where("status IN ?", []string{
			runtimestate.RuntimeManagedInstanceStatusCreating,
			runtimestate.RuntimeManagedInstanceStatusRunning,
			runtimestate.RuntimeManagedInstanceStatusStopping,
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
			runtimestate.RuntimeManagedInstanceStatusCreating,
			runtimestate.RuntimeManagedInstanceStatusRunning,
			runtimestate.RuntimeManagedInstanceStatusStopping,
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
