package infrastructure

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

func (r *Repository) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*runtimecontracts.RuntimeManagedInstance, error) {
	var instance runtimecontracts.RuntimeManagedInstance
	err := r.dbWithContext(ctx).Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?", userID, challengeID,
		[]string{
			runtimecontracts.RuntimeManagedInstanceStatusPending,
			runtimecontracts.RuntimeManagedInstanceStatusCreating,
			runtimecontracts.RuntimeManagedInstanceStatusRunning,
		}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	return r.dbWithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"expires_at": expiresAt,
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *Repository) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	_, err := r.updateStatusAndReleasePortWithCurrentStatus(ctx, id, nil, status)
	return err
}

func (r *Repository) FailProvisioning(ctx context.Context, id int64) (bool, error) {
	return r.updateStatusAndReleasePortWithCurrentStatus(
		ctx,
		id,
		[]string{runtimecontracts.RuntimeManagedInstanceStatusCreating},
		runtimecontracts.RuntimeManagedInstanceStatusFailed,
	)
}

func (r *Repository) updateStatusAndReleasePortWithCurrentStatus(ctx context.Context, id int64, currentStatuses []string, status string) (bool, error) {
	if id <= 0 {
		return false, nil
	}

	changed := false
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var instance runtimecontracts.RuntimeManagedInstance
		query := tx.Select("id", "host_port").Where("id = ?", id)
		if len(currentStatuses) > 0 {
			query = query.Where("status IN ?", currentStatuses)
		}
		if err := query.First(&instance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && len(currentStatuses) > 0 {
				return nil
			}
			return err
		}

		now := time.Now().UTC()
		updates := map[string]any{
			"status":     status,
			"updated_at": now,
		}
		if status == runtimecontracts.RuntimeManagedInstanceStatusStopped || status == runtimecontracts.RuntimeManagedInstanceStatusExpired {
			updates["destroyed_at"] = now
			updates["host_port"] = 0
			updates["container_id"] = ""
			updates["network_id"] = ""
			updates["runtime_details"] = ""
			updates["access_url"] = ""
		}
		updateQuery := tx.Model(&runtimecontracts.RuntimeManagedInstance{}).
			Where("id = ?", id)
		if len(currentStatuses) > 0 {
			updateQuery = updateQuery.Where("status IN ?", currentStatuses)
		}
		result := updateQuery.Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if len(currentStatuses) > 0 && result.RowsAffected == 0 {
			return nil
		}

		deleteQuery := tx.Where("instance_id = ?", id)
		if instance.HostPort > 0 {
			deleteQuery = deleteQuery.Or("port = ?", instance.HostPort)
		}
		if err := deleteQuery.Delete(&runtimeentity.PortAllocation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", id).Delete(&runtimeentity.NetworkAllocation{}).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	return changed, err
}

func (r *Repository) FinalizeStoppedRuntime(ctx context.Context, id int64) error {
	return r.finalizeInstanceRuntime(ctx, id, runtimecontracts.RuntimeManagedInstanceStatusStopped)
}

func (r *Repository) ExpireInstanceRuntime(ctx context.Context, id int64) error {
	return r.finalizeInstanceRuntime(ctx, id, runtimecontracts.RuntimeManagedInstanceStatusExpired)
}

func (r *Repository) finalizeInstanceRuntime(ctx context.Context, id int64, status string) error {
	if id <= 0 {
		return nil
	}

	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var instance runtimecontracts.RuntimeManagedInstance
		if err := tx.Select("id", "host_port").Where("id = ?", id).First(&instance).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		if err := tx.Model(&runtimecontracts.RuntimeManagedInstance{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"status":          status,
				"host_port":       0,
				"container_id":    "",
				"network_id":      "",
				"runtime_details": "",
				"access_url":      "",
				"destroyed_at":    now,
				"updated_at":      now,
			}).Error; err != nil {
			return err
		}

		deleteQuery := tx.Where("instance_id = ?", id)
		if instance.HostPort > 0 {
			deleteQuery = deleteQuery.Or("port = ?", instance.HostPort)
		}
		if err := deleteQuery.Delete(&runtimeentity.PortAllocation{}).Error; err != nil {
			return err
		}
		return tx.Where("instance_id = ?", id).Delete(&runtimeentity.NetworkAllocation{}).Error
	})
}

func (r *Repository) UpdateRuntime(ctx context.Context, instance *runtimecontracts.RuntimeManagedInstance) error {
	_, err := r.PersistProvisionedRuntime(ctx, instance)
	return err
}

func (r *Repository) PersistProvisionedRuntime(ctx context.Context, instance *runtimecontracts.RuntimeManagedInstance) (bool, error) {
	if instance == nil || instance.ID <= 0 {
		return false, nil
	}
	result := r.dbWithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("id = ? AND status = ?", instance.ID, runtimecontracts.RuntimeManagedInstanceStatusCreating).
		Updates(map[string]any{
			"contest_id":      instance.ContestID,
			"team_id":         instance.TeamID,
			"host_port":       instance.HostPort,
			"container_id":    instance.ContainerID,
			"network_id":      instance.NetworkID,
			"runtime_details": instance.RuntimeDetails,
			"access_url":      instance.AccessURL,
			"status":          instance.Status,
			"updated_at":      time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
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

func (r *Repository) FindExpired(ctx context.Context) ([]*runtimecontracts.RuntimeManagedInstance, error) {
	var instances []*runtimecontracts.RuntimeManagedInstance
	err := r.dbWithContext(ctx).Where("status = ? AND expires_at < ?",
		runtimecontracts.RuntimeManagedInstanceStatusRunning, time.Now().UTC()).
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListRecoverableActiveInstances(ctx context.Context) ([]*runtimecontracts.RuntimeManagedInstance, error) {
	var instances []*runtimecontracts.RuntimeManagedInstance
	err := r.dbWithContext(ctx).
		Where("status IN ?", []string{
			runtimecontracts.RuntimeManagedInstanceStatusCreating,
			runtimecontracts.RuntimeManagedInstanceStatusRunning,
		}).
		Where("expires_at > ?", time.Now().UTC()).
		Order("updated_at ASC, id ASC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListStoppingInstances(ctx context.Context, updatedBefore time.Time, limit int) ([]*runtimecontracts.RuntimeManagedInstance, error) {
	var instances []*runtimecontracts.RuntimeManagedInstance
	query := r.dbWithContext(ctx).
		Where("status = ?", runtimecontracts.RuntimeManagedInstanceStatusStopping)
	if !updatedBefore.IsZero() {
		query = query.Where("updated_at <= ?", updatedBefore)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("updated_at ASC, id ASC").Find(&instances).Error
	return instances, err
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

func (r *Repository) RefreshActiveAWDInstanceExpiryByContest(ctx context.Context, contestID int64, activeAt, expiresAt time.Time) error {
	if contestID <= 0 || expiresAt.IsZero() {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("contest_id = ? AND service_id IS NOT NULL AND status IN ?", contestID, []string{
			runtimecontracts.RuntimeManagedInstanceStatusPending,
			runtimecontracts.RuntimeManagedInstanceStatusCreating,
			runtimecontracts.RuntimeManagedInstanceStatusRunning,
		}).
		Where("expires_at > ?", activeAt.UTC()).
		Updates(map[string]any{
			"expires_at": expiresAt.UTC(),
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *Repository) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}

	result := r.dbWithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("id = ? AND status IN ? AND expires_at > ?",
			id,
			[]string{
				runtimecontracts.RuntimeManagedInstanceStatusCreating,
				runtimecontracts.RuntimeManagedInstanceStatusRunning,
			},
			time.Now().UTC(),
		).
		Updates(map[string]any{
			"status":          runtimecontracts.RuntimeManagedInstanceStatusPending,
			"container_id":    "",
			"network_id":      "",
			"runtime_details": "",
			"access_url":      "",
			"updated_at":      time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
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

func (r *Repository) ListPendingInstances(ctx context.Context, limit int) ([]*runtimecontracts.RuntimeManagedInstance, error) {
	if limit <= 0 {
		return []*runtimecontracts.RuntimeManagedInstance{}, nil
	}

	instances := make([]*runtimecontracts.RuntimeManagedInstance, 0, limit)
	err := r.db.WithContext(ctx).
		Where("status = ?", runtimecontracts.RuntimeManagedInstanceStatusPending).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *Repository) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]any{
			"status":     toStatus,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("status IN ?", statuses).
		Count(&count).Error
	return count, err
}

func (r *Repository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	return r.ReserveAvailablePortExcluding(ctx, start, end, 0)
}

func (r *Repository) ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error) {
	for port := start; port < end; port++ {
		if excludedPort > 0 && port == excludedPort {
			continue
		}
		reserved, err := r.tryReservePort(ctx, port)
		if err != nil {
			return 0, err
		}
		if reserved {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port in range %d-%d", start, end)
}

func (r *Repository) tryReservePort(ctx context.Context, port int) (bool, error) {
	now := time.Now().UTC()
	result := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "port"}},
		DoUpdates: clause.Assignments(map[string]any{
			"updated_at": now,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "port_allocations.instance_id IS NULL"},
		}},
	}).Create(&runtimeentity.PortAllocation{
		Port:      port,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if result.Error != nil {
		if isPortAllocationConflict(result.Error) {
			return false, nil
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	return r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).
		Where("port = ?", port).
		Updates(map[string]any{
			"instance_id": instanceID,
			"updated_at":  time.Now().UTC(),
		}).Error
}

func (r *Repository) ReleaseReservedPort(ctx context.Context, port int) error {
	if port <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("port = ? AND instance_id IS NULL", port).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *Repository) ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error {
	if port <= 0 || instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("port = ? AND instance_id = ?", port, instanceID).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *Repository) ReserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int) (string, error) {
	return r.ReserveAvailableSubnetExcluding(ctx, baseCIDR, subnetMask, nil)
}

func (r *Repository) ReserveAvailableSubnetForInstance(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string) (string, error) {
	return r.ReserveAvailableSubnetForInstanceExcluding(ctx, baseCIDR, subnetMask, instanceID, networkKey, nil)
}

func (r *Repository) ReserveAvailableSubnetExcluding(ctx context.Context, baseCIDR string, subnetMask int, excludedSubnets []string) (string, error) {
	return r.reserveAvailableSubnet(ctx, baseCIDR, subnetMask, 0, "", excludedSubnets)
}

func (r *Repository) ReserveAvailableSubnetForInstanceExcluding(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
	return r.reserveAvailableSubnet(ctx, baseCIDR, subnetMask, instanceID, networkKey, excludedSubnets)
}

func (r *Repository) reserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
	normalizedKey := strings.TrimSpace(networkKey)
	excludedSet := make(map[string]struct{}, len(excludedSubnets))
	for _, subnet := range excludedSubnets {
		subnet = strings.TrimSpace(subnet)
		if subnet == "" {
			continue
		}
		excludedSet[subnet] = struct{}{}
	}
	if instanceID > 0 && normalizedKey != "" {
		existing, err := r.findSubnetAllocationByOwner(ctx, instanceID, normalizedKey)
		if err != nil {
			return "", err
		}
		if existing != "" {
			if _, excluded := excludedSet[existing]; excluded {
				existing = ""
			}
		}
		if existing != "" {
			return existing, nil
		}
	}

	candidates, err := subnetCandidates(baseCIDR, subnetMask)
	if err != nil {
		return "", err
	}
	for _, subnet := range candidates {
		if _, excluded := excludedSet[subnet]; excluded {
			continue
		}
		reserved, reserveErr := r.tryReserveSubnet(ctx, subnet, instanceID, normalizedKey)
		if reserveErr != nil {
			return "", reserveErr
		}
		if reserved {
			return subnet, nil
		}
		if instanceID > 0 && normalizedKey != "" {
			existing, findErr := r.findSubnetAllocationByOwner(ctx, instanceID, normalizedKey)
			if findErr != nil {
				return "", findErr
			}
			if existing != "" {
				if _, excluded := excludedSet[existing]; excluded {
					moved, moveErr := r.moveSubnetAllocationForOwner(ctx, instanceID, normalizedKey, existing, subnet)
					if moveErr != nil {
						return "", moveErr
					}
					if moved {
						return subnet, nil
					}
					continue
				}
				return existing, nil
			}
		}
	}
	return "", fmt.Errorf("no available subnet in %s with /%d", baseCIDR, subnetMask)
}

func (r *Repository) tryReserveSubnet(ctx context.Context, subnet string, instanceID int64, networkKey string) (bool, error) {
	now := time.Now().UTC()
	allocation := &runtimeentity.NetworkAllocation{
		Subnet:     subnet,
		NetworkKey: networkKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if instanceID > 0 {
		allocation.InstanceID = &instanceID
	}

	result := r.dbWithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(allocation)
	if result.Error != nil {
		if isNetworkAllocationConflict(result.Error) {
			return false, nil
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) moveSubnetAllocationForOwner(ctx context.Context, instanceID int64, networkKey, currentSubnet, targetSubnet string) (bool, error) {
	if instanceID <= 0 || strings.TrimSpace(networkKey) == "" {
		return false, nil
	}
	currentSubnet = strings.TrimSpace(currentSubnet)
	targetSubnet = strings.TrimSpace(targetSubnet)
	if currentSubnet == "" || targetSubnet == "" || currentSubnet == targetSubnet {
		return false, nil
	}

	result := r.dbWithContext(ctx).
		Model(&runtimeentity.NetworkAllocation{}).
		Where("instance_id = ? AND network_key = ? AND subnet = ?", instanceID, networkKey, currentSubnet).
		Updates(map[string]any{
			"subnet":     targetSubnet,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		if isNetworkAllocationConflict(result.Error) {
			return false, nil
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) findSubnetAllocationByOwner(ctx context.Context, instanceID int64, networkKey string) (string, error) {
	if instanceID <= 0 || strings.TrimSpace(networkKey) == "" {
		return "", nil
	}

	var allocation runtimeentity.NetworkAllocation
	err := r.dbWithContext(ctx).
		Where("instance_id = ? AND network_key = ?", instanceID, networkKey).
		First(&allocation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return allocation.Subnet, nil
}

func (r *Repository) ReleaseReservedSubnet(ctx context.Context, subnet string) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("subnet = ? AND instance_id IS NULL", subnet).
		Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *Repository) ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" || instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("subnet = ? AND instance_id = ?", subnet, instanceID).
		Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *Repository) IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
	if instanceID <= 0 || hostPort <= 0 {
		return false, nil
	}

	var allocation runtimeentity.PortAllocation
	if err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("port = ?", hostPort).
		First(&allocation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if allocation.InstanceID == nil {
		return false, nil
	}
	return *allocation.InstanceID == instanceID, nil
}

func (r *Repository) SyncInstanceHostPortForRestart(ctx context.Context, instanceID int64, hostPort int, preserveHostPort bool) (int, error) {
	if instanceID <= 0 {
		return 0, nil
	}
	if !preserveHostPort {
		return 0, r.releaseAllPortsForInstance(ctx, instanceID)
	}

	boundPort, err := r.findLatestBoundPortForInstance(ctx, instanceID)
	if err != nil {
		return 0, err
	}
	if boundPort > 0 {
		hostPort = boundPort
	}
	if hostPort <= 0 {
		return 0, nil
	}
	if err := r.ensurePortBoundToInstance(ctx, hostPort, instanceID); err != nil {
		return 0, err
	}
	return hostPort, nil
}

func (r *Repository) findLatestBoundPortForInstance(ctx context.Context, instanceID int64) (int, error) {
	if instanceID <= 0 {
		return 0, nil
	}

	var allocation runtimeentity.PortAllocation
	err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("instance_id = ?", instanceID).
		Order("updated_at DESC, port DESC").
		First(&allocation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return allocation.Port, nil
}

func (r *Repository) ensurePortBoundToInstance(ctx context.Context, port int, instanceID int64) error {
	if port <= 0 || instanceID <= 0 {
		return nil
	}

	allocation := &runtimeentity.PortAllocation{
		Port:       port,
		InstanceID: &instanceID,
	}
	if err := r.dbWithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(allocation).Error; err != nil {
		return err
	}

	var stored runtimeentity.PortAllocation
	if err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("port = ?", port).
		First(&stored).Error; err != nil {
		return err
	}
	if stored.InstanceID != nil && *stored.InstanceID != instanceID {
		return fmt.Errorf("host port %d is allocated to instance %d", port, *stored.InstanceID)
	}
	if stored.InstanceID == nil {
		return r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).
			Where("port = ?", port).
			Updates(map[string]any{
				"instance_id": instanceID,
				"updated_at":  time.Now().UTC(),
			}).Error
	}
	return nil
}

func (r *Repository) releaseAllPortsForInstance(ctx context.Context, instanceID int64) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("instance_id = ?", instanceID).
		Delete(&runtimeentity.PortAllocation{}).Error
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

func (r *Repository) ListAllocatedPorts(ctx context.Context) ([]int, error) {
	var ports []int
	if err := r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).Pluck("port", &ports).Error; err == nil {
		return ports, nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "no such table") && !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		return nil, err
	}

	var accessURLs []string
	if err := r.dbWithContext(ctx).Model(&runtimecontracts.RuntimeManagedInstance{}).
		Where("status IN ?", []string{
			runtimecontracts.RuntimeManagedInstanceStatusCreating,
			runtimecontracts.RuntimeManagedInstanceStatusRunning,
		}).
		Where("access_url <> ''").
		Pluck("access_url", &accessURLs).Error; err != nil {
		return nil, err
	}

	ports = make([]int, 0, len(accessURLs))
	for _, rawURL := range accessURLs {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		portValue := parsed.Port()
		if portValue == "" {
			continue
		}
		port, err := strconv.Atoi(portValue)
		if err != nil {
			continue
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func isPortAllocationConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "unique constraint failed") ||
		strings.Contains(lowered, "duplicate key value") ||
		strings.Contains(lowered, "duplicate entry")
}

func isNetworkAllocationConflict(err error) bool {
	return isPortAllocationConflict(err)
}

func subnetCandidates(baseCIDR string, subnetMask int) ([]string, error) {
	baseIP, baseNet, err := net.ParseCIDR(strings.TrimSpace(baseCIDR))
	if err != nil {
		return nil, fmt.Errorf("parse subnet base: %w", err)
	}
	ip4 := baseIP.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("subnet base must be ipv4")
	}
	baseNet.IP = ip4

	basePrefix, bits := baseNet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("subnet base must be ipv4")
	}
	if subnetMask <= basePrefix || subnetMask > 30 {
		return nil, fmt.Errorf("invalid subnet mask /%d for base %s", subnetMask, baseCIDR)
	}

	start := binary.BigEndian.Uint32(baseNet.IP)
	blockSize := uint32(1) << uint32(32-subnetMask)
	subnetCount := 1 << uint(subnetMask-basePrefix)
	result := make([]string, 0, subnetCount)
	for idx := 0; idx < subnetCount; idx++ {
		current := start + uint32(idx)*blockSize
		ip := make(net.IP, net.IPv4len)
		binary.BigEndian.PutUint32(ip, current)
		result = append(result, (&net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(subnetMask, 32),
		}).String())
	}
	return result, nil
}
