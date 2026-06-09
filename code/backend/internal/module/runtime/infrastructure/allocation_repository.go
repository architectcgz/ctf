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

type AllocationRepository struct {
	db *gorm.DB
}

func NewAllocationRepository(db *gorm.DB) *AllocationRepository {
	return &AllocationRepository{db: db}
}

func (r *AllocationRepository) WithDB(db *gorm.DB) *AllocationRepository {
	return &AllocationRepository{db: db}
}

func (r *AllocationRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *AllocationRepository) ReleaseRuntimeAllocationsForInstance(ctx context.Context, instanceID int64, hostPort int) error {
	if instanceID <= 0 {
		return nil
	}

	deleteQuery := r.dbWithContext(ctx).Where("instance_id = ?", instanceID)
	if hostPort > 0 {
		deleteQuery = deleteQuery.Or("port = ?", hostPort)
	}
	if err := deleteQuery.Delete(&runtimeentity.PortAllocation{}).Error; err != nil {
		return err
	}
	return r.dbWithContext(ctx).Where("instance_id = ?", instanceID).Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *AllocationRepository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	return r.ReserveAvailablePortExcluding(ctx, start, end, 0)
}

func (r *AllocationRepository) ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error) {
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

func (r *AllocationRepository) tryReservePort(ctx context.Context, port int) (bool, error) {
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

func (r *AllocationRepository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	return r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).
		Where("port = ?", port).
		Updates(map[string]any{
			"instance_id": instanceID,
			"updated_at":  time.Now().UTC(),
		}).Error
}

func (r *AllocationRepository) ReleaseReservedPort(ctx context.Context, port int) error {
	if port <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("port = ? AND instance_id IS NULL", port).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *AllocationRepository) ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error {
	if port <= 0 || instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("port = ? AND instance_id = ?", port, instanceID).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *AllocationRepository) ReserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int) (string, error) {
	return r.ReserveAvailableSubnetExcluding(ctx, baseCIDR, subnetMask, nil)
}

func (r *AllocationRepository) ReserveAvailableSubnetForInstance(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string) (string, error) {
	return r.ReserveAvailableSubnetForInstanceExcluding(ctx, baseCIDR, subnetMask, instanceID, networkKey, nil)
}

func (r *AllocationRepository) ReserveAvailableSubnetExcluding(ctx context.Context, baseCIDR string, subnetMask int, excludedSubnets []string) (string, error) {
	return r.reserveAvailableSubnet(ctx, baseCIDR, subnetMask, 0, "", excludedSubnets)
}

func (r *AllocationRepository) ReserveAvailableSubnetForInstanceExcluding(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
	return r.reserveAvailableSubnet(ctx, baseCIDR, subnetMask, instanceID, networkKey, excludedSubnets)
}

func (r *AllocationRepository) reserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
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

func (r *AllocationRepository) tryReserveSubnet(ctx context.Context, subnet string, instanceID int64, networkKey string) (bool, error) {
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

func (r *AllocationRepository) moveSubnetAllocationForOwner(ctx context.Context, instanceID int64, networkKey, currentSubnet, targetSubnet string) (bool, error) {
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

func (r *AllocationRepository) findSubnetAllocationByOwner(ctx context.Context, instanceID int64, networkKey string) (string, error) {
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

func (r *AllocationRepository) ReleaseReservedSubnet(ctx context.Context, subnet string) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("subnet = ? AND instance_id IS NULL", subnet).
		Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *AllocationRepository) ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" || instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("subnet = ? AND instance_id = ?", subnet, instanceID).
		Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *AllocationRepository) IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
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

func (r *AllocationRepository) SyncInstanceHostPortForRestart(ctx context.Context, instanceID int64, hostPort int, preserveHostPort bool) (int, error) {
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

func (r *AllocationRepository) findLatestBoundPortForInstance(ctx context.Context, instanceID int64) (int, error) {
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

func (r *AllocationRepository) ensurePortBoundToInstance(ctx context.Context, port int, instanceID int64) error {
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

func (r *AllocationRepository) releaseAllPortsForInstance(ctx context.Context, instanceID int64) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("instance_id = ?", instanceID).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *AllocationRepository) ListAllocatedPorts(ctx context.Context) ([]int, error) {
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
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, current)
		result = append(result, (&net.IPNet{IP: ip, Mask: net.CIDRMask(subnetMask, 32)}).String())
	}
	return result, nil
}
