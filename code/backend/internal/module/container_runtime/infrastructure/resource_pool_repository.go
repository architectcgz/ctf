package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ctf-platform/internal/config"
	runtimeentity "ctf-platform/internal/module/container_runtime/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RuntimeResourcePoolRepository struct {
	db *gorm.DB
}

const runtimeResourcePoolSeedBatchSize = 500

func NewRuntimeResourcePoolRepository(db *gorm.DB) *RuntimeResourcePoolRepository {
	if db == nil {
		return nil
	}
	return &RuntimeResourcePoolRepository{db: db}
}

func (r *RuntimeResourcePoolRepository) dbWithContext(ctx context.Context) *gorm.DB {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx)
}

func (r *RuntimeResourcePoolRepository) EnsurePoolsForNode(ctx context.Context, nodeID int64, cfg config.ContainerConfig) error {
	if r == nil || r.db == nil || nodeID <= 0 {
		return nil
	}
	now := time.Now().UTC()
	portRows := make([]runtimeentity.RuntimePortPool, 0)
	if cfg.PortRangeEnd > cfg.PortRangeStart {
		portRows = make([]runtimeentity.RuntimePortPool, 0, cfg.PortRangeEnd-cfg.PortRangeStart)
	}
	for port := cfg.PortRangeStart; port < cfg.PortRangeEnd; port++ {
		portRows = append(portRows, runtimeentity.RuntimePortPool{
			RuntimeNodeID: nodeID,
			Port:          port,
			Status:        runtimeentity.RuntimeResourceStatusAvailable,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}
	if err := r.createRuntimePortPoolRows(ctx, portRows); err != nil {
		return err
	}
	if err := r.ensureSubnetPool(ctx, nodeID, runtimeentity.RuntimeSubnetPoolKindSingleContainer, cfg.Network.SingleContainerSubnetBase, cfg.Network.SingleContainerSubnetMask, now); err != nil {
		return err
	}
	return r.ensureSubnetPool(ctx, nodeID, runtimeentity.RuntimeSubnetPoolKindTopology, cfg.Network.TopologySubnetBase, cfg.Network.TopologySubnetMask, now)
}

func (r *RuntimeResourcePoolRepository) ensureSubnetPool(ctx context.Context, nodeID int64, poolKind, baseCIDR string, subnetMask int, now time.Time) error {
	poolKind = strings.TrimSpace(poolKind)
	if nodeID <= 0 || poolKind == "" || strings.TrimSpace(baseCIDR) == "" || subnetMask <= 0 {
		return nil
	}
	candidates, err := subnetCandidates(baseCIDR, subnetMask)
	if err != nil {
		return err
	}
	rows := make([]runtimeentity.RuntimeSubnetPool, 0, len(candidates))
	for _, subnet := range candidates {
		rows = append(rows, runtimeentity.RuntimeSubnetPool{
			RuntimeNodeID: nodeID,
			PoolKind:      poolKind,
			Subnet:        subnet,
			Status:        runtimeentity.RuntimeResourceStatusAvailable,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}
	return r.createRuntimeSubnetPoolRows(ctx, rows)
}

func (r *RuntimeResourcePoolRepository) createRuntimePortPoolRows(ctx context.Context, rows []runtimeentity.RuntimePortPool) error {
	if len(rows) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(rows, runtimeResourcePoolSeedBatchSize).Error
}

func (r *RuntimeResourcePoolRepository) createRuntimeSubnetPoolRows(ctx context.Context, rows []runtimeentity.RuntimeSubnetPool) error {
	if len(rows) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(rows, runtimeResourcePoolSeedBatchSize).Error
}

func (r *RuntimeResourcePoolRepository) ReserveAvailablePortForNode(ctx context.Context, nodeID, instanceID int64) (int, error) {
	if r == nil || r.db == nil || nodeID <= 0 {
		return 0, fmt.Errorf("runtime node resource pool is unavailable")
	}
	var reservedPort int
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row runtimeentity.RuntimePortPool
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("runtime_node_id = ? AND status = ?", nodeID, runtimeentity.RuntimeResourceStatusAvailable).
			Order("port ASC").
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("no available port for runtime node %d", nodeID)
			}
			return err
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"status":      runtimeentity.RuntimeResourceStatusReserved,
			"reserved_at": now,
			"updated_at":  now,
		}
		if instanceID > 0 {
			updates["instance_id"] = instanceID
		}
		result := tx.Model(&runtimeentity.RuntimePortPool{}).
			Where("runtime_node_id = ? AND port = ? AND status = ?", row.RuntimeNodeID, row.Port, runtimeentity.RuntimeResourceStatusAvailable).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("port %d for runtime node %d is no longer available", row.Port, row.RuntimeNodeID)
		}
		reservedPort = row.Port
		return nil
	})
	return reservedPort, err
}

func (r *RuntimeResourcePoolRepository) ReserveAvailableSubnetForNode(ctx context.Context, nodeID int64, poolKind string, instanceID int64, networkKey string) (string, error) {
	if r == nil || r.db == nil || nodeID <= 0 {
		return "", fmt.Errorf("runtime node resource pool is unavailable")
	}
	poolKind = strings.TrimSpace(poolKind)
	if poolKind == "" {
		return "", fmt.Errorf("runtime subnet pool kind is required")
	}
	var reservedSubnet string
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row runtimeentity.RuntimeSubnetPool
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("runtime_node_id = ? AND pool_kind = ? AND status = ?", nodeID, poolKind, runtimeentity.RuntimeResourceStatusAvailable).
			Order("subnet ASC").
			First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("no available %s subnet for runtime node %d", poolKind, nodeID)
			}
			return err
		}
		now := time.Now().UTC()
		updates := map[string]any{
			"status":      runtimeentity.RuntimeResourceStatusReserved,
			"network_key": strings.TrimSpace(networkKey),
			"reserved_at": now,
			"updated_at":  now,
		}
		if instanceID > 0 {
			updates["instance_id"] = instanceID
		}
		result := tx.Model(&runtimeentity.RuntimeSubnetPool{}).
			Where("runtime_node_id = ? AND subnet = ? AND status = ?", row.RuntimeNodeID, row.Subnet, runtimeentity.RuntimeResourceStatusAvailable).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("subnet %s for runtime node %d is no longer available", row.Subnet, row.RuntimeNodeID)
		}
		reservedSubnet = row.Subnet
		return nil
	})
	return reservedSubnet, err
}

func (r *RuntimeResourcePoolRepository) BindResourcesForInstance(ctx context.Context, instanceID int64) error {
	if r == nil || r.db == nil || instanceID <= 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&runtimeentity.RuntimePortPool{}).
			Where("instance_id = ? AND status = ?", instanceID, runtimeentity.RuntimeResourceStatusReserved).
			Updates(map[string]any{"status": runtimeentity.RuntimeResourceStatusBound, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&runtimeentity.RuntimeSubnetPool{}).
			Where("instance_id = ? AND status = ?", instanceID, runtimeentity.RuntimeResourceStatusReserved).
			Updates(map[string]any{"status": runtimeentity.RuntimeResourceStatusBound, "updated_at": now}).Error
	})
}

func (r *RuntimeResourcePoolRepository) ReleaseResourcesForInstance(ctx context.Context, instanceID int64) error {
	if r == nil || r.db == nil || instanceID <= 0 {
		return nil
	}
	now := time.Now().UTC()
	release := map[string]any{
		"status":      runtimeentity.RuntimeResourceStatusAvailable,
		"instance_id": nil,
		"network_key": "",
		"reserved_at": nil,
		"updated_at":  now,
	}
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&runtimeentity.RuntimePortPool{}).
			Where("instance_id = ?", instanceID).
			Updates(map[string]any{
				"status":      runtimeentity.RuntimeResourceStatusAvailable,
				"instance_id": nil,
				"reserved_at": nil,
				"updated_at":  now,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&runtimeentity.RuntimeSubnetPool{}).
			Where("instance_id = ?", instanceID).
			Updates(release).Error
	})
}

func (r *RuntimeResourcePoolRepository) QuarantinePort(ctx context.Context, nodeID int64, port int, reason string) error {
	_ = strings.TrimSpace(reason)
	if r == nil || r.db == nil || nodeID <= 0 || port <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).Model(&runtimeentity.RuntimePortPool{}).
		Where("runtime_node_id = ? AND port = ?", nodeID, port).
		Updates(map[string]any{
			"status":      runtimeentity.RuntimeResourceStatusQuarantined,
			"instance_id": nil,
			"reserved_at": nil,
			"updated_at":  time.Now().UTC(),
		}).Error
}

func (r *RuntimeResourcePoolRepository) QuarantineSubnet(ctx context.Context, nodeID int64, subnet string, reason string) error {
	_ = strings.TrimSpace(reason)
	subnet = strings.TrimSpace(subnet)
	if r == nil || r.db == nil || nodeID <= 0 || subnet == "" {
		return nil
	}
	return r.dbWithContext(ctx).Model(&runtimeentity.RuntimeSubnetPool{}).
		Where("runtime_node_id = ? AND subnet = ?", nodeID, subnet).
		Updates(map[string]any{
			"status":      runtimeentity.RuntimeResourceStatusQuarantined,
			"instance_id": nil,
			"reserved_at": nil,
			"updated_at":  time.Now().UTC(),
		}).Error
}
