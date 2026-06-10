package infrastructure

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimestate "ctf-platform/internal/module/runtime/contracts"
)

type ContainerNodeIndexRepository struct {
	db *gorm.DB
}

func NewContainerNodeIndexRepository(db *gorm.DB) *ContainerNodeIndexRepository {
	return &ContainerNodeIndexRepository{db: db}
}

func (r *ContainerNodeIndexRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ContainerNodeIndexRepository) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
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
		Model(&runtimestate.RuntimeManagedInstance{}).
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
