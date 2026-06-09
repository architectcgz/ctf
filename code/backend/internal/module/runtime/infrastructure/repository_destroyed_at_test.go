package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

func TestReserveAvailablePortExcludingSkipsExcludedPort(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)

	port, err := repo.ReserveAvailablePortExcluding(context.Background(), 30000, 30003, 30000)
	if err != nil {
		t.Fatalf("ReserveAvailablePortExcluding() error = %v", err)
	}
	if port != 30001 {
		t.Fatalf("expected excluded port 30000 to be skipped, got %d", port)
	}
}

func TestReserveAvailablePortExcludingReusesUnboundAllocation(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)

	staleUpdatedAt := time.Now().Add(-time.Hour).UTC()
	if err := db.Create(&runtimeentity.PortAllocation{
		Port:      30000,
		CreatedAt: staleUpdatedAt,
		UpdatedAt: staleUpdatedAt,
	}).Error; err != nil {
		t.Fatalf("seed unbound port allocation: %v", err)
	}

	port, err := repo.ReserveAvailablePortExcluding(context.Background(), 30000, 30003, 0)
	if err != nil {
		t.Fatalf("ReserveAvailablePortExcluding() error = %v", err)
	}
	if port != 30000 {
		t.Fatalf("expected stale unbound allocation on port 30000 to be reused, got %d", port)
	}

	var allocations []runtimeentity.PortAllocation
	if err := db.Order("port ASC").Find(&allocations).Error; err != nil {
		t.Fatalf("load port allocations: %v", err)
	}
	if len(allocations) != 1 {
		t.Fatalf("expected one port allocation row after reuse, got %d", len(allocations))
	}
	if allocations[0].Port != 30000 {
		t.Fatalf("expected reused allocation to stay on port 30000, got %+v", allocations[0])
	}
	if allocations[0].InstanceID != nil {
		t.Fatalf("expected reused allocation to remain unbound before bind, got %+v", allocations[0].InstanceID)
	}
	if !allocations[0].UpdatedAt.After(staleUpdatedAt) {
		t.Fatalf("expected reused allocation updated_at to advance, got %v <= %v", allocations[0].UpdatedAt, staleUpdatedAt)
	}
}

func TestReserveAvailableSubnetForInstanceSkipsAllocatedSubnet(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)
	otherInstanceID := int64(9001)
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.0.0/24",
		InstanceID: &otherInstanceID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
	}).Error; err != nil {
		t.Fatalf("seed subnet allocation: %v", err)
	}

	subnet, err := repo.ReserveAvailableSubnetForInstance(context.Background(), "10.10.0.0/16", 24, 9002, runtimecontracts.TopologyDefaultNetworkKey)
	if err != nil {
		t.Fatalf("ReserveAvailableSubnetForInstance() error = %v", err)
	}
	if subnet != "10.10.1.0/24" {
		t.Fatalf("expected next available subnet 10.10.1.0/24, got %q", subnet)
	}
}

func TestReserveAvailableSubnetForInstanceReusesOwnerReservation(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)
	instanceID := int64(9101)
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.9.0/24",
		InstanceID: &instanceID,
		NetworkKey: "backend",
	}).Error; err != nil {
		t.Fatalf("seed subnet allocation: %v", err)
	}

	subnet, err := repo.ReserveAvailableSubnetForInstance(context.Background(), "10.10.0.0/16", 24, instanceID, "backend")
	if err != nil {
		t.Fatalf("ReserveAvailableSubnetForInstance() error = %v", err)
	}
	if subnet != "10.10.9.0/24" {
		t.Fatalf("expected existing owner subnet 10.10.9.0/24, got %q", subnet)
	}
}

func TestReserveAvailableSubnetForInstanceExcludingSkipsExcludedSubnet(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)

	subnet, err := repo.ReserveAvailableSubnetForInstanceExcluding(
		context.Background(),
		"10.10.0.0/16",
		24,
		9201,
		runtimecontracts.TopologyDefaultNetworkKey,
		[]string{"10.10.0.0/24", "10.10.1.0/24"},
	)
	if err != nil {
		t.Fatalf("ReserveAvailableSubnetForInstanceExcluding() error = %v", err)
	}
	if subnet != "10.10.2.0/24" {
		t.Fatalf("expected excluded subnets to be skipped, got %q", subnet)
	}
}

func TestReserveAvailableSubnetForInstanceExcludingReassignsExcludedOwnerReservation(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)
	instanceID := int64(9301)
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.9.0/24",
		InstanceID: &instanceID,
		NetworkKey: "backend",
	}).Error; err != nil {
		t.Fatalf("seed subnet allocation: %v", err)
	}

	subnet, err := repo.ReserveAvailableSubnetForInstanceExcluding(
		context.Background(),
		"10.10.0.0/16",
		24,
		instanceID,
		"backend",
		[]string{"10.10.9.0/24"},
	)
	if err != nil {
		t.Fatalf("ReserveAvailableSubnetForInstanceExcluding() error = %v", err)
	}
	if subnet != "10.10.0.0/24" {
		t.Fatalf("expected owner reservation to move to first available subnet, got %q", subnet)
	}

	var allocation runtimeentity.NetworkAllocation
	if err := db.Where("instance_id = ? AND network_key = ?", instanceID, "backend").First(&allocation).Error; err != nil {
		t.Fatalf("load updated subnet allocation: %v", err)
	}
	if allocation.Subnet != "10.10.0.0/24" {
		t.Fatalf("expected owner allocation to update to 10.10.0.0/24, got %q", allocation.Subnet)
	}
}

func TestSyncInstanceHostPortForRestartPreservesAndBindsAllocation(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	repo := NewRepository(db)

	instanceID := int64(6001)
	if err := db.Create(&instancecontracts.Instance{
		ID:          instanceID,
		UserID:      7,
		ChallengeID: 11,
		HostPort:    32021,
		Status:      instancecontracts.InstanceStatusFailed,
		ExpiresAt:   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&runtimeentity.PortAllocation{Port: 32021}).Error; err != nil {
		t.Fatalf("seed unbound port allocation: %v", err)
	}

	hostPort, err := repo.SyncInstanceHostPortForRestart(context.Background(), instanceID, 32021, true)
	if err != nil {
		t.Fatalf("SyncInstanceHostPortForRestart() error = %v", err)
	}
	if hostPort != 32021 {
		t.Fatalf("expected preserved host port 32021, got %d", hostPort)
	}

	var allocation runtimeentity.PortAllocation
	if err := db.Where("port = ?", 32021).First(&allocation).Error; err != nil {
		t.Fatalf("load port allocation: %v", err)
	}
	if allocation.InstanceID == nil || *allocation.InstanceID != instanceID {
		t.Fatalf("expected port allocation to bind to instance %d, got %+v", instanceID, allocation.InstanceID)
	}
}

func TestUpdateStatusAndReleasePortSetsDestroyedAtForStoppedInstance(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	instance := instancecontracts.Instance{
		ID:             1,
		UserID:         7,
		ChallengeID:    11,
		ContainerID:    "inst-running",
		NetworkID:      "net-running",
		RuntimeDetails: `{"containers":[{"container_id":"inst-running","host_port":32001}]}`,
		HostPort:       32001,
		Status:         instancecontracts.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:32001",
		CreatedAt:      now.Add(-30 * time.Minute),
		UpdatedAt:      now.Add(-10 * time.Minute),
		ExpiresAt:      now.Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&runtimeentity.PortAllocation{Port: instance.HostPort, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("seed port allocation: %v", err)
	}
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.7.0/24",
		InstanceID: &instance.ID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
	}).Error; err != nil {
		t.Fatalf("seed network allocation: %v", err)
	}

	before := time.Now()
	if err := updateStatusAndReleasePort(context.Background(), db, instance.ID, instancecontracts.InstanceStatusStopped); err != nil {
		t.Fatalf("UpdateStatusAndReleasePort() error = %v", err)
	}
	after := time.Now()

	var row struct {
		Status         string     `gorm:"column:status"`
		HostPort       int        `gorm:"column:host_port"`
		ContainerID    string     `gorm:"column:container_id"`
		NetworkID      string     `gorm:"column:network_id"`
		RuntimeDetails string     `gorm:"column:runtime_details"`
		AccessURL      string     `gorm:"column:access_url"`
		DestroyedAt    *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").
		Select("status", "host_port", "container_id", "network_id", "runtime_details", "access_url", "destroyed_at").
		Where("id = ?", instance.ID).
		Take(&row).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusStopped {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusStopped)
	}
	if row.HostPort != 0 || row.ContainerID != "" || row.NetworkID != "" || row.RuntimeDetails != "" || row.AccessURL != "" {
		t.Fatalf("expected runtime fields to be cleared, got %+v", row)
	}
	if row.DestroyedAt == nil {
		t.Fatal("expected destroyed_at to be set for stopped instance")
	}
	if row.DestroyedAt.Before(before.Add(-time.Second)) || row.DestroyedAt.After(after.Add(time.Second)) {
		t.Fatalf("destroyed_at = %v, want between %v and %v", row.DestroyedAt, before, after)
	}

	var remaining int64
	if err := db.Model(&runtimeentity.PortAllocation{}).Where("instance_id = ? OR port = ?", instance.ID, instance.HostPort).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining port allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected port allocations to be released, got %d", remaining)
	}
	if err := db.Model(&runtimeentity.NetworkAllocation{}).Where("instance_id = ?", instance.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining network allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected network allocations to be released, got %d", remaining)
	}
}

func TestUpdateStatusAndReleasePortClearsRuntimeFieldsForExpiredInstance(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	now := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)
	instance := instancecontracts.Instance{
		ID:             2,
		UserID:         8,
		ChallengeID:    12,
		ContainerID:    "inst-expiring",
		NetworkID:      "net-expiring",
		RuntimeDetails: `{"containers":[{"container_id":"inst-expiring","host_port":32002}]}`,
		HostPort:       32002,
		Status:         instancecontracts.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:32002",
		CreatedAt:      now.Add(-30 * time.Minute),
		UpdatedAt:      now.Add(-10 * time.Minute),
		ExpiresAt:      now.Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&runtimeentity.PortAllocation{Port: instance.HostPort, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("seed port allocation: %v", err)
	}
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.8.0/24",
		InstanceID: &instance.ID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
	}).Error; err != nil {
		t.Fatalf("seed network allocation: %v", err)
	}

	before := time.Now()
	if err := updateStatusAndReleasePort(context.Background(), db, instance.ID, instancecontracts.InstanceStatusExpired); err != nil {
		t.Fatalf("UpdateStatusAndReleasePort() error = %v", err)
	}
	after := time.Now()

	var row struct {
		Status         string     `gorm:"column:status"`
		HostPort       int        `gorm:"column:host_port"`
		ContainerID    string     `gorm:"column:container_id"`
		NetworkID      string     `gorm:"column:network_id"`
		RuntimeDetails string     `gorm:"column:runtime_details"`
		AccessURL      string     `gorm:"column:access_url"`
		DestroyedAt    *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").
		Select("status", "host_port", "container_id", "network_id", "runtime_details", "access_url", "destroyed_at").
		Where("id = ?", instance.ID).
		Take(&row).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusExpired {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusExpired)
	}
	if row.HostPort != 0 || row.ContainerID != "" || row.NetworkID != "" || row.RuntimeDetails != "" || row.AccessURL != "" {
		t.Fatalf("expected runtime fields to be cleared, got %+v", row)
	}
	if row.DestroyedAt == nil {
		t.Fatal("expected destroyed_at to be set for expired instance")
	}
	if row.DestroyedAt.Before(before.Add(-time.Second)) || row.DestroyedAt.After(after.Add(time.Second)) {
		t.Fatalf("destroyed_at = %v, want between %v and %v", row.DestroyedAt, before, after)
	}

	var remaining int64
	if err := db.Model(&runtimeentity.PortAllocation{}).Where("instance_id = ? OR port = ?", instance.ID, instance.HostPort).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining port allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected port allocations to be released, got %d", remaining)
	}
	if err := db.Model(&runtimeentity.NetworkAllocation{}).Where("instance_id = ?", instance.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining network allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected network allocations to be released, got %d", remaining)
	}
}

func TestFailProvisioningDoesNotOverrideStoppingInstance(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)

	instance := instancecontracts.Instance{
		ID:          12,
		UserID:      7,
		ChallengeID: 100,
		Status:      instancecontracts.InstanceStatusStopping,
		HostPort:    32012,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&runtimeentity.PortAllocation{Port: instance.HostPort, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("seed port allocation: %v", err)
	}

	changed, err := failProvisioning(context.Background(), db, instance.ID)
	if err != nil {
		t.Fatalf("FailProvisioning() error = %v", err)
	}
	if changed {
		t.Fatal("expected stopping instance to reject fail-provisioning update")
	}

	var row struct {
		Status string `gorm:"column:status"`
	}
	if err := db.Table("instances").Select("status").Where("id = ?", instance.ID).Take(&row).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusStopping {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusStopping)
	}

	var remaining int64
	if err := db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", instance.HostPort).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining port allocations: %v", err)
	}
	if remaining != 1 {
		t.Fatalf("expected fail-provisioning rejection to preserve port allocation, got %d", remaining)
	}
}

func TestUpdateStatusAndReleasePortDoesNotSetDestroyedAtForFailedInstance(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	instance := instancecontracts.Instance{
		ID:          2,
		UserID:      9,
		ChallengeID: 15,
		ContainerID: "inst-creating",
		HostPort:    32002,
		Status:      instancecontracts.InstanceStatusCreating,
		CreatedAt:   time.Now().Add(-5 * time.Minute),
		UpdatedAt:   time.Now().Add(-2 * time.Minute),
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}

	if err := updateStatusAndReleasePort(context.Background(), db, instance.ID, instancecontracts.InstanceStatusFailed); err != nil {
		t.Fatalf("UpdateStatusAndReleasePort() error = %v", err)
	}

	var row struct {
		Status      string     `gorm:"column:status"`
		DestroyedAt *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").Select("status", "destroyed_at").Where("id = ?", instance.ID).Take(&row).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusFailed {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusFailed)
	}
	if row.DestroyedAt != nil {
		t.Fatalf("expected destroyed_at to stay nil for failed instance, got %v", row.DestroyedAt)
	}
}

func TestExpireInstanceRuntimeClearsRuntimeFieldsAndPortAllocation(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	now := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
	instance := instancecontracts.Instance{
		ID:             3,
		UserID:         10,
		ChallengeID:    16,
		HostPort:       32003,
		ContainerID:    "inst-runtime",
		NetworkID:      "net-runtime",
		RuntimeDetails: `{"containers":[{"container_id":"inst-runtime","host_port":32003}]}`,
		Status:         instancecontracts.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:32003",
		CreatedAt:      now.Add(-5 * time.Minute),
		UpdatedAt:      now.Add(-2 * time.Minute),
		ExpiresAt:      now.Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&runtimeentity.PortAllocation{Port: instance.HostPort, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("seed port allocation: %v", err)
	}
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.8.0/24",
		InstanceID: &instance.ID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
	}).Error; err != nil {
		t.Fatalf("seed network allocation: %v", err)
	}

	before := time.Now()
	if err := expireInstanceRuntime(context.Background(), db, instance.ID); err != nil {
		t.Fatalf("ExpireInstanceRuntime() error = %v", err)
	}
	after := time.Now()

	var row struct {
		Status         string     `gorm:"column:status"`
		HostPort       int        `gorm:"column:host_port"`
		ContainerID    string     `gorm:"column:container_id"`
		NetworkID      string     `gorm:"column:network_id"`
		RuntimeDetails string     `gorm:"column:runtime_details"`
		AccessURL      string     `gorm:"column:access_url"`
		DestroyedAt    *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").
		Select("status", "host_port", "container_id", "network_id", "runtime_details", "access_url", "destroyed_at").
		Where("id = ?", instance.ID).
		Take(&row).Error; err != nil {
		t.Fatalf("load expired instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusExpired {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusExpired)
	}
	if row.HostPort != 0 || row.ContainerID != "" || row.NetworkID != "" || row.RuntimeDetails != "" || row.AccessURL != "" {
		t.Fatalf("expected runtime fields to be cleared, got %+v", row)
	}
	if row.DestroyedAt == nil {
		t.Fatal("expected destroyed_at to be set for expired instance")
	}
	if row.DestroyedAt.Before(before.Add(-time.Second)) || row.DestroyedAt.After(after.Add(time.Second)) {
		t.Fatalf("destroyed_at = %v, want between %v and %v", row.DestroyedAt, before, after)
	}

	var remaining int64
	if err := db.Model(&runtimeentity.PortAllocation{}).Where("instance_id = ? OR port = ?", instance.ID, instance.HostPort).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining port allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected port allocations to be released, got %d", remaining)
	}
	if err := db.Model(&runtimeentity.NetworkAllocation{}).Where("instance_id = ?", instance.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining network allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected network allocations to be released, got %d", remaining)
	}
}

func TestFinalizeStoppedRuntimeClearsRuntimeFieldsAndAllocations(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	instance := instancecontracts.Instance{
		ID:             31,
		UserID:         10,
		ChallengeID:    16,
		HostPort:       32031,
		ContainerID:    "inst-runtime",
		NetworkID:      "net-runtime",
		RuntimeDetails: `{"containers":[{"container_id":"inst-runtime","host_port":32031}]}`,
		Status:         instancecontracts.InstanceStatusStopping,
		AccessURL:      "http://127.0.0.1:32031",
		CreatedAt:      now.Add(-5 * time.Minute),
		UpdatedAt:      now.Add(-2 * time.Minute),
		ExpiresAt:      now.Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&runtimeentity.PortAllocation{Port: instance.HostPort, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("seed port allocation: %v", err)
	}
	if err := db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.31.0/24",
		InstanceID: &instance.ID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
	}).Error; err != nil {
		t.Fatalf("seed network allocation: %v", err)
	}

	before := time.Now()
	if err := finalizeStoppedRuntime(context.Background(), db, instance.ID); err != nil {
		t.Fatalf("FinalizeStoppedRuntime() error = %v", err)
	}
	after := time.Now()

	var row struct {
		Status         string     `gorm:"column:status"`
		HostPort       int        `gorm:"column:host_port"`
		ContainerID    string     `gorm:"column:container_id"`
		NetworkID      string     `gorm:"column:network_id"`
		RuntimeDetails string     `gorm:"column:runtime_details"`
		AccessURL      string     `gorm:"column:access_url"`
		DestroyedAt    *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").
		Select("status", "host_port", "container_id", "network_id", "runtime_details", "access_url", "destroyed_at").
		Where("id = ?", instance.ID).
		Take(&row).Error; err != nil {
		t.Fatalf("load stopped instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusStopped {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusStopped)
	}
	if row.HostPort != 0 || row.ContainerID != "" || row.NetworkID != "" || row.RuntimeDetails != "" || row.AccessURL != "" {
		t.Fatalf("expected runtime fields to be cleared, got %+v", row)
	}
	if row.DestroyedAt == nil {
		t.Fatal("expected destroyed_at to be set for stopped instance")
	}
	if row.DestroyedAt.Before(before.Add(-time.Second)) || row.DestroyedAt.After(after.Add(time.Second)) {
		t.Fatalf("destroyed_at = %v, want between %v and %v", row.DestroyedAt, before, after)
	}

	var remaining int64
	if err := db.Model(&runtimeentity.PortAllocation{}).Where("instance_id = ? OR port = ?", instance.ID, instance.HostPort).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining port allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected port allocations to be released, got %d", remaining)
	}
	if err := db.Model(&runtimeentity.NetworkAllocation{}).Where("instance_id = ?", instance.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining network allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected network allocations to be released, got %d", remaining)
	}
}

func newRuntimeRepositoryDestroyedAtTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&instancecontracts.Instance{}, &runtimeentity.PortAllocation{}, &runtimeentity.NetworkAllocation{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}

func updateStatusAndReleasePort(ctx context.Context, db *gorm.DB, id int64, status string) error {
	return runRuntimeLifecycleTx(ctx, db, func(instanceTx *instanceinfra.Repository, runtimeTx *Repository) error {
		release, err := instanceTx.UpdateStatus(ctx, id, status)
		if err != nil || release == nil {
			return err
		}
		return runtimeTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func failProvisioning(ctx context.Context, db *gorm.DB, id int64) (bool, error) {
	changed := false
	err := runRuntimeLifecycleTx(ctx, db, func(instanceTx *instanceinfra.Repository, runtimeTx *Repository) error {
		release, failed, err := instanceTx.FailProvisioning(ctx, id)
		if err != nil {
			return err
		}
		changed = failed
		if !failed || release == nil {
			return nil
		}
		return runtimeTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
	return changed, err
}

func expireInstanceRuntime(ctx context.Context, db *gorm.DB, id int64) error {
	return runRuntimeLifecycleTx(ctx, db, func(instanceTx *instanceinfra.Repository, runtimeTx *Repository) error {
		release, err := instanceTx.ExpireInstanceRuntime(ctx, id)
		if err != nil || release == nil {
			return err
		}
		return runtimeTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func finalizeStoppedRuntime(ctx context.Context, db *gorm.DB, id int64) error {
	return runRuntimeLifecycleTx(ctx, db, func(instanceTx *instanceinfra.Repository, runtimeTx *Repository) error {
		release, err := instanceTx.FinalizeStoppedRuntime(ctx, id)
		if err != nil || release == nil {
			return err
		}
		return runtimeTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func runRuntimeLifecycleTx(ctx context.Context, db *gorm.DB, fn func(instanceTx *instanceinfra.Repository, runtimeTx *Repository) error) error {
	if db == nil || fn == nil {
		return nil
	}
	instanceRepo := instanceinfra.NewRepository(db)
	runtimeRepo := NewRepository(db)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(instanceRepo.WithDB(tx), runtimeRepo.WithDB(tx))
	})
}
