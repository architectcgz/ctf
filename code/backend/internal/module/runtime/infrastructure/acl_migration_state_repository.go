package infrastructure

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
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

func (r *ACLMigrationStateRepository) ListInstancesNeedingACLHandleMigration(ctx context.Context) ([]runtimecontracts.RuntimeManagedInstance, error) {
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

func (r *ACLMigrationStateRepository) UpdateInstanceRuntimeDetails(ctx context.Context, instanceID int64, runtimeDetails string) error {
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
