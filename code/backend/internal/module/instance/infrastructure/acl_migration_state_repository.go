package infrastructure

import (
	"context"
	"strings"
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"

	"gorm.io/gorm"
)

type ACLMigrationStateRepository struct {
	db *gorm.DB
}

func NewACLMigrationStateRepository(db *gorm.DB) *ACLMigrationStateRepository {
	return &ACLMigrationStateRepository{db: db}
}

func (r *ACLMigrationStateRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ACLMigrationStateRepository) ListInstancesNeedingACLHandleMigration(ctx context.Context) ([]instancecontracts.Instance, error) {
	type instanceACLMigrationRow struct {
		ID             int64  `gorm:"column:id"`
		NodeID         *int64 `gorm:"column:node_id"`
		RuntimeDetails string `gorm:"column:runtime_details"`
	}

	rows := make([]instanceACLMigrationRow, 0)
	if err := r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
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

	result := make([]instancecontracts.Instance, 0, len(rows))
	for _, row := range rows {
		result = append(result, instancecontracts.Instance{
			ID:             row.ID,
			NodeID:         row.NodeID,
			RuntimeDetails: row.RuntimeDetails,
		})
	}
	return result, nil
}

func (r *ACLMigrationStateRepository) UpdateInstanceRuntimeDetails(ctx context.Context, instanceID int64, runtimeDetails string) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"runtime_details": runtimeDetails,
			"updated_at":      time.Now().UTC(),
		}).Error
}
