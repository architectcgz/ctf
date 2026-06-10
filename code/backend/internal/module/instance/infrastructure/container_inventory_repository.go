package infrastructure

import (
	"context"
	"strings"

	instancecontracts "ctf-platform/internal/module/instance/contracts"

	"gorm.io/gorm"
)

type ContainerInventoryRepository struct {
	db *gorm.DB
}

func NewContainerInventoryRepository(db *gorm.DB) *ContainerInventoryRepository {
	return &ContainerInventoryRepository{db: db}
}

func (r *ContainerInventoryRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ContainerInventoryRepository) ListActiveContainerInventory(ctx context.Context) ([]instancecontracts.Instance, error) {
	items := make([]instancecontracts.Instance, 0)
	if err := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status IN ?", []string{
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusStopping,
		}).
		Select("id, node_id, container_id, runtime_details").
		Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ContainerInventoryRepository) ListContainerNodeLookupCandidates(ctx context.Context, containerID string) ([]instancecontracts.Instance, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, nil
	}

	rows := make([]instancecontracts.Instance, 0)
	likePattern := "%" + containerID + "%"
	if err := r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Select("id, node_id, container_id, runtime_details").
		Where("container_id = ? OR runtime_details LIKE ?", containerID, likePattern).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
