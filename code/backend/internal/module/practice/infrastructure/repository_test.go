package infrastructure_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	containerruntimeentity "ctf-platform/internal/module/container_runtime/entity"
	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

func newRepositoryWithRuntimePortOwner(db *gorm.DB) *practiceinfra.Repository {
	return practiceinfra.NewRepositoryWithRuntimePortOwner(db, func(db *gorm.DB) runtimeports.PortReservationOwner {
		return containerruntimeinfra.NewAllocationRepository(db)
	})
}

func TestRepositoryReserveAvailablePortSkipsAllocatedPort(t *testing.T) {
	db := newRepositoryTestDB(t, &containerruntimeentity.PortAllocation{})
	ownerInstanceID := int64(400)
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30000, InstanceID: &ownerInstanceID}).Error; err != nil {
		t.Fatalf("seed allocated port: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	port, err := repo.ReserveAvailablePort(context.Background(), 30000, 30002)
	if err != nil {
		t.Fatalf("ReserveAvailablePort() error = %v", err)
	}
	if port != 30001 {
		t.Fatalf("expected port 30001, got %d", port)
	}

	var count int64
	if err := db.Model(&containerruntimeentity.PortAllocation{}).Where("port IN ?", []int{30000, 30001}).Count(&count).Error; err != nil {
		t.Fatalf("count allocated ports: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected two allocated ports, got %d", count)
	}
}

func TestRepositoryReserveAvailablePortExcludingSkipsExcludedPort(t *testing.T) {
	db := newRepositoryTestDB(t, &containerruntimeentity.PortAllocation{})

	repo := newRepositoryWithRuntimePortOwner(db)
	port, err := repo.ReserveAvailablePortExcluding(context.Background(), 30000, 30003, 30000)
	if err != nil {
		t.Fatalf("ReserveAvailablePortExcluding() error = %v", err)
	}
	if port != 30001 {
		t.Fatalf("expected excluded port 30000 to be skipped, got %d", port)
	}
}

func TestRepositoryReleasePortForInstanceOnlyDeletesOwnedAllocation(t *testing.T) {
	db := newRepositoryTestDB(t, &containerruntimeentity.PortAllocation{})

	ownerInstanceID := int64(401)
	otherInstanceID := int64(402)
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30015, InstanceID: &ownerInstanceID}).Error; err != nil {
		t.Fatalf("seed allocated port: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ReleasePortForInstance(context.Background(), 30015, otherInstanceID); err != nil {
		t.Fatalf("ReleasePortForInstance() with foreign owner error = %v", err)
	}

	var count int64
	if err := db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30015).Count(&count).Error; err != nil {
		t.Fatalf("count preserved allocation: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected foreign release to keep allocation, got %d rows", count)
	}

	if err := repo.ReleasePortForInstance(context.Background(), 30015, ownerInstanceID); err != nil {
		t.Fatalf("ReleasePortForInstance() with owner error = %v", err)
	}
	if err := db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30015).Count(&count).Error; err != nil {
		t.Fatalf("count released allocation: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected owner release to delete allocation, got %d rows", count)
	}
}

func TestRepositoryReleaseReservedPortOnlyDeletesUnboundAllocation(t *testing.T) {
	db := newRepositoryTestDB(t, &containerruntimeentity.PortAllocation{})

	ownerInstanceID := int64(501)
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30016}).Error; err != nil {
		t.Fatalf("seed unbound port allocation: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30017, InstanceID: &ownerInstanceID}).Error; err != nil {
		t.Fatalf("seed bound port allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ReleaseReservedPort(context.Background(), 30017); err != nil {
		t.Fatalf("ReleaseReservedPort() with bound allocation error = %v", err)
	}

	var count int64
	if err := db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30017).Count(&count).Error; err != nil {
		t.Fatalf("count bound allocation: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected bound allocation to stay, got %d rows", count)
	}

	if err := repo.ReleaseReservedPort(context.Background(), 30016); err != nil {
		t.Fatalf("ReleaseReservedPort() with unbound allocation error = %v", err)
	}
	if err := db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30016).Count(&count).Error; err != nil {
		t.Fatalf("count unbound allocation: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected unbound allocation to be deleted, got %d rows", count)
	}
}

func TestRepositoryCreateAWDServiceOperationClosesStaleActiveScopeEntries(t *testing.T) {
	db := newRepositoryTestDB(t, &runtimeentity.AWDServiceOperation{})

	startedAt := time.Date(2026, 6, 5, 3, 3, 38, 0, time.UTC)
	stale := &runtimeentity.AWDServiceOperation{
		ContestID:     8,
		TeamID:        15,
		ServiceID:     22,
		InstanceID:    2379,
		OperationType: runtimeentity.AWDServiceOperationTypeStart,
		RequestedBy:   runtimeentity.AWDServiceOperationRequestedBySystem,
		Reason:        "desired_runtime_reconcile",
		SLABillable:   false,
		Status:        runtimeentity.AWDServiceOperationStatusProvisioning,
		StartedAt:     startedAt.Add(-5 * time.Minute),
		CreatedAt:     startedAt.Add(-5 * time.Minute),
		UpdatedAt:     startedAt.Add(-5 * time.Minute),
	}
	if err := db.Create(stale).Error; err != nil {
		t.Fatalf("seed stale awd operation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	replacement := &runtimeentity.AWDServiceOperation{
		ContestID:     8,
		TeamID:        15,
		ServiceID:     22,
		InstanceID:    2540,
		OperationType: runtimeentity.AWDServiceOperationTypeRecreate,
		RequestedBy:   runtimeentity.AWDServiceOperationRequestedBySystem,
		Reason:        "desired_runtime_reconcile",
		SLABillable:   false,
		Status:        runtimeentity.AWDServiceOperationStatusProvisioning,
		StartedAt:     startedAt,
		CreatedAt:     startedAt,
		UpdatedAt:     startedAt,
	}
	if err := repo.CreateAWDServiceOperation(context.Background(), replacement); err != nil {
		t.Fatalf("CreateAWDServiceOperation() error = %v", err)
	}

	var operations []runtimeentity.AWDServiceOperation
	if err := db.Order("id ASC").Find(&operations).Error; err != nil {
		t.Fatalf("query awd operations: %v", err)
	}
	if len(operations) != 2 {
		t.Fatalf("expected stale and replacement operations, got %+v", operations)
	}

	closed := operations[0]
	if closed.Status != runtimeentity.AWDServiceOperationStatusFailed {
		t.Fatalf("expected stale operation to be closed as failed, got %+v", closed)
	}
	if closed.FinishedAt == nil || !closed.FinishedAt.Equal(startedAt) {
		t.Fatalf("expected stale operation finished at replacement start time %s, got %+v", startedAt, closed)
	}
	if closed.ErrorMessage != "superseded_by_new_operation" {
		t.Fatalf("expected stale operation superseded marker, got %+v", closed)
	}

	active := operations[1]
	if active.Status != runtimeentity.AWDServiceOperationStatusProvisioning {
		t.Fatalf("expected replacement operation to stay active, got %+v", active)
	}
	if active.InstanceID != replacement.InstanceID {
		t.Fatalf("expected replacement instance %d, got %+v", replacement.InstanceID, active)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartClearsHostPortWhenNotPreserved(t *testing.T) {
	db := newRepositoryTestDB(t, &instancecontracts.Instance{}, &containerruntimeentity.PortAllocation{})

	otherInstanceID := int64(98)
	instance := instancecontracts.Instance{
		ID:          99,
		UserID:      3,
		ChallengeID: 4,
		HostPort:    30000,
		Status:      instancecontracts.InstanceStatusFailed,
		ShareScope:  instancecontracts.ShareScopePerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30000, InstanceID: &otherInstanceID}).Error; err != nil {
		t.Fatalf("seed other allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, instancecontracts.InstanceStatusPending, time.Now().Add(2*time.Hour), false); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored instancecontracts.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 0 || stored.Status != instancecontracts.InstanceStatusPending {
		t.Fatalf("expected host port cleared and pending status, got host_port=%d status=%s", stored.HostPort, stored.Status)
	}

	var allocation containerruntimeentity.PortAllocation
	if err := db.First(&allocation, "port = ?", 30000).Error; err != nil {
		t.Fatalf("expected other allocation to remain: %v", err)
	}
	if allocation.InstanceID == nil || *allocation.InstanceID != otherInstanceID {
		t.Fatalf("expected allocation to stay with instance %d, got %+v", otherInstanceID, allocation.InstanceID)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartReleasesOwnedHostPortWhenNotPreserved(t *testing.T) {
	db := newRepositoryTestDB(t, &instancecontracts.Instance{}, &containerruntimeentity.PortAllocation{})

	instanceID := int64(100)
	instance := instancecontracts.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 5,
		HostPort:    30002,
		Status:      instancecontracts.InstanceStatusFailed,
		ShareScope:  instancecontracts.ShareScopePerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30002, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, instancecontracts.InstanceStatusPending, time.Now().Add(2*time.Hour), false); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var count int64
	if err := db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30002).Count(&count).Error; err != nil {
		t.Fatalf("count allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected owned host port allocation to be released, got %d rows", count)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartPreservesOwnedHostPort(t *testing.T) {
	db := newRepositoryTestDB(t, &instancecontracts.Instance{}, &containerruntimeentity.PortAllocation{})

	instanceID := int64(101)
	instance := instancecontracts.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 6,
		HostPort:    30001,
		Status:      instancecontracts.InstanceStatusRunning,
		ShareScope:  instancecontracts.ShareScopePerUser,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30001, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, instancecontracts.InstanceStatusPending, time.Now().Add(2*time.Hour), true); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored instancecontracts.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 30001 {
		t.Fatalf("expected host port preserved, got %d", stored.HostPort)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartSyncsBoundAllocationWhenHostPortMissing(t *testing.T) {
	db := newRepositoryTestDB(t, &instancecontracts.Instance{}, &containerruntimeentity.PortAllocation{})

	instanceID := int64(102)
	instance := instancecontracts.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 7,
		HostPort:    0,
		Status:      instancecontracts.InstanceStatusFailed,
		ShareScope:  instancecontracts.ShareScopePerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30003, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, instancecontracts.InstanceStatusPending, time.Now().Add(2*time.Hour), true); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored instancecontracts.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 30003 {
		t.Fatalf("expected host port to sync from bound allocation, got %d", stored.HostPort)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartUsesBoundAllocationWhenStoredHostPortConflicts(t *testing.T) {
	db := newRepositoryTestDB(t, &instancecontracts.Instance{}, &containerruntimeentity.PortAllocation{})

	instanceID := int64(103)
	otherInstanceID := int64(104)
	instance := instancecontracts.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 8,
		HostPort:    30004,
		Status:      instancecontracts.InstanceStatusFailed,
		ShareScope:  instancecontracts.ShareScopePerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30004, InstanceID: &otherInstanceID}).Error; err != nil {
		t.Fatalf("seed conflicting allocation: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30007, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed rebound allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, instancecontracts.InstanceStatusPending, time.Now().Add(2*time.Hour), true); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored instancecontracts.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 30007 {
		t.Fatalf("expected host port to switch to rebound allocation, got %d", stored.HostPort)
	}
}

func TestRepositoryFindContestAWDServiceRuntimeSubjectMapsSnapshot(t *testing.T) {
	db := newRepositoryTestDB(t, &contestentity.ContestAWDService{})

	service := &contestentity.ContestAWDService{
		ID:              41,
		ContestID:       7,
		AWDChallengeID:  19,
		DisplayName:     "Display Name",
		IsVisible:       true,
		ScoreConfig:     `{"points":320}`,
		ServiceSnapshot: `{"name":"Snapshot Name","category":"web","difficulty":"medium","runtime_config":{"image_id":105,"instance_sharing":"per_team","topology":{"entry_node_key":"web","spec":{"nodes":[{"key":"web"}]}},"defense_workspace":{"seed_root":"docker/workspace","workspace_roots":["docker/workspace/src"],"writable_roots":["docker/workspace/src"],"runtime_mounts":[{"source":"docker/workspace/src","target":"/workspace/src","mode":"rw"}]}},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
	}
	if err := db.Create(service).Error; err != nil {
		t.Fatalf("seed awd service: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)
	subject, err := repo.FindContestAWDServiceRuntimeSubject(context.Background(), service.ContestID, service.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceRuntimeSubject() error = %v", err)
	}
	if subject == nil {
		t.Fatal("expected runtime subject")
	}
	if subject.ServiceID != service.ID || subject.ChallengeID != service.AWDChallengeID || !subject.Visible {
		t.Fatalf("unexpected runtime subject identity: %+v", subject)
	}
	if subject.RuntimeChallenge == nil {
		t.Fatal("expected runtime challenge")
	}
	if subject.RuntimeChallenge.Title != "Display Name" {
		t.Fatalf("expected display name to win title fallback, got %+v", subject.RuntimeChallenge)
	}
	if subject.RuntimeChallenge.Points != 320 || subject.RuntimeChallenge.ImageID != 105 {
		t.Fatalf("unexpected runtime challenge payload: %+v", subject.RuntimeChallenge)
	}
	if subject.RuntimeChallenge.InstanceSharing != string(challengecontracts.InstanceSharingPerTeam) {
		t.Fatalf("unexpected instance sharing: %+v", subject.RuntimeChallenge)
	}
	if subject.SeedSignature == "" {
		t.Fatalf("expected seed signature, got %+v", subject)
	}
	if subject.RuntimeTopology == nil || subject.RuntimeTopology.EntryNodeKey != "web" {
		t.Fatalf("unexpected runtime topology: %+v", subject.RuntimeTopology)
	}
	if subject.RuntimeTopology.Spec == "" {
		t.Fatalf("expected topology spec, got %+v", subject.RuntimeTopology)
	}
	if subject.WorkspaceConfig == nil {
		t.Fatalf("expected workspace config, got %+v", subject)
	}
	if subject.WorkspaceConfig.SeedRoot != "docker/workspace" {
		t.Fatalf("unexpected workspace config: %+v", subject.WorkspaceConfig)
	}
	if subject.WorkspaceConfig.CheckerTokenEnv != "" {
		t.Fatalf("expected empty checker token env by default, got %+v", subject.WorkspaceConfig)
	}
}

func TestRepositoryIsHostPortReusableForRestart(t *testing.T) {
	db := newRepositoryTestDB(t, &containerruntimeentity.PortAllocation{})

	currentInstanceID := int64(201)
	otherInstanceID := int64(202)
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30011, InstanceID: &currentInstanceID}).Error; err != nil {
		t.Fatalf("seed owned allocation: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30012, InstanceID: &otherInstanceID}).Error; err != nil {
		t.Fatalf("seed foreign allocation: %v", err)
	}
	if err := db.Create(&containerruntimeentity.PortAllocation{Port: 30013}).Error; err != nil {
		t.Fatalf("seed unbound allocation: %v", err)
	}

	repo := newRepositoryWithRuntimePortOwner(db)

	reusable, err := repo.IsHostPortReusableForRestart(context.Background(), currentInstanceID, 30011)
	if err != nil {
		t.Fatalf("owned IsHostPortReusableForRestart() error = %v", err)
	}
	if !reusable {
		t.Fatal("expected owned host port to be reusable")
	}

	reusable, err = repo.IsHostPortReusableForRestart(context.Background(), currentInstanceID, 30012)
	if err != nil {
		t.Fatalf("foreign IsHostPortReusableForRestart() error = %v", err)
	}
	if reusable {
		t.Fatal("expected foreign-owned host port to be rejected")
	}

	reusable, err = repo.IsHostPortReusableForRestart(context.Background(), currentInstanceID, 30013)
	if err != nil {
		t.Fatalf("unbound IsHostPortReusableForRestart() error = %v", err)
	}
	if reusable {
		t.Fatal("expected unbound host port allocation to be rejected")
	}

	reusable, err = repo.IsHostPortReusableForRestart(context.Background(), currentInstanceID, 30014)
	if err != nil {
		t.Fatalf("missing IsHostPortReusableForRestart() error = %v", err)
	}
	if reusable {
		t.Fatal("expected missing host port allocation to be rejected")
	}
}

func newRepositoryTestDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return db
}
