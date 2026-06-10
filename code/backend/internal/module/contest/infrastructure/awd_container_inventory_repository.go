package infrastructure

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type AWDContainerInventoryRepository struct {
	db *gorm.DB
}

func NewAWDContainerInventoryRepository(db *gorm.DB) *AWDContainerInventoryRepository {
	return &AWDContainerInventoryRepository{db: db}
}

func (r *AWDContainerInventoryRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *AWDContainerInventoryRepository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	var items []struct {
		ContainerID string `gorm:"column:container_id"`
	}
	if err := r.dbWithContext(ctx).
		Table("awd_defense_workspaces AS ws").
		Joins("JOIN instances AS inst ON inst.id = ws.instance_id").
		Where("inst.status IN ?", []string{
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusStopping,
		}).
		Where("ws.status = ? AND ws.container_id <> ''", contestentity.AWDDefenseWorkspaceStatusRunning).
		Select("ws.container_id").
		Scan(&items).Error; err != nil {
		if isMissingAWDContainerInventoryTable(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
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

func (r *AWDContainerInventoryRepository) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, nil
	}

	var workspace struct {
		NodeID *int64 `gorm:"column:node_id"`
	}
	if err := r.dbWithContext(ctx).
		Table("awd_defense_workspaces AS ws").
		Joins("JOIN instances AS inst ON inst.id = ws.instance_id").
		Where("ws.container_id = ?", containerID).
		Select("inst.node_id").
		Take(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || isMissingAWDContainerInventoryTable(err) {
			return nil, nil
		}
		return nil, err
	}
	return workspace.NodeID, nil
}

func isMissingAWDContainerInventoryTable(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist")
}
