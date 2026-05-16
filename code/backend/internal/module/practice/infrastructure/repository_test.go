package infrastructure_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ctf-platform/internal/model"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
)

func TestRepositoryReserveAvailablePortSkipsAllocatedPort(t *testing.T) {
	db := newRepositoryTestDB(t, &model.PortAllocation{})
	if err := db.Create(&model.PortAllocation{Port: 30000}).Error; err != nil {
		t.Fatalf("seed allocated port: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	port, err := repo.ReserveAvailablePort(context.Background(), 30000, 30002)
	if err != nil {
		t.Fatalf("ReserveAvailablePort() error = %v", err)
	}
	if port != 30001 {
		t.Fatalf("expected port 30001, got %d", port)
	}

	var count int64
	if err := db.Model(&model.PortAllocation{}).Where("port IN ?", []int{30000, 30001}).Count(&count).Error; err != nil {
		t.Fatalf("count allocated ports: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected two allocated ports, got %d", count)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartClearsHostPortWhenNotPreserved(t *testing.T) {
	db := newRepositoryTestDB(t, &model.Instance{}, &model.PortAllocation{})

	otherInstanceID := int64(98)
	instance := model.Instance{
		ID:          99,
		UserID:      3,
		ChallengeID: 4,
		HostPort:    30000,
		Status:      model.InstanceStatusFailed,
		ShareScope:  model.InstanceSharingPerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30000, InstanceID: &otherInstanceID}).Error; err != nil {
		t.Fatalf("seed other allocation: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, model.InstanceStatusPending, time.Now().Add(2*time.Hour), false); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored model.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 0 || stored.Status != model.InstanceStatusPending {
		t.Fatalf("expected host port cleared and pending status, got host_port=%d status=%s", stored.HostPort, stored.Status)
	}

	var allocation model.PortAllocation
	if err := db.First(&allocation, "port = ?", 30000).Error; err != nil {
		t.Fatalf("expected other allocation to remain: %v", err)
	}
	if allocation.InstanceID == nil || *allocation.InstanceID != otherInstanceID {
		t.Fatalf("expected allocation to stay with instance %d, got %+v", otherInstanceID, allocation.InstanceID)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartReleasesOwnedHostPortWhenNotPreserved(t *testing.T) {
	db := newRepositoryTestDB(t, &model.Instance{}, &model.PortAllocation{})

	instanceID := int64(100)
	instance := model.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 5,
		HostPort:    30002,
		Status:      model.InstanceStatusFailed,
		ShareScope:  model.InstanceSharingPerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30002, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, model.InstanceStatusPending, time.Now().Add(2*time.Hour), false); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var count int64
	if err := db.Model(&model.PortAllocation{}).Where("port = ?", 30002).Count(&count).Error; err != nil {
		t.Fatalf("count allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected owned host port allocation to be released, got %d rows", count)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartPreservesOwnedHostPort(t *testing.T) {
	db := newRepositoryTestDB(t, &model.Instance{}, &model.PortAllocation{})

	instanceID := int64(101)
	instance := model.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 6,
		HostPort:    30001,
		Status:      model.InstanceStatusRunning,
		ShareScope:  model.InstanceSharingPerUser,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30001, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, model.InstanceStatusPending, time.Now().Add(2*time.Hour), true); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored model.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 30001 {
		t.Fatalf("expected host port preserved, got %d", stored.HostPort)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartSyncsBoundAllocationWhenHostPortMissing(t *testing.T) {
	db := newRepositoryTestDB(t, &model.Instance{}, &model.PortAllocation{})

	instanceID := int64(102)
	instance := model.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 7,
		HostPort:    0,
		Status:      model.InstanceStatusFailed,
		ShareScope:  model.InstanceSharingPerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30003, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, model.InstanceStatusPending, time.Now().Add(2*time.Hour), true); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored model.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 30003 {
		t.Fatalf("expected host port to sync from bound allocation, got %d", stored.HostPort)
	}
}

func TestRepositoryResetInstanceRuntimeForRestartUsesBoundAllocationWhenStoredHostPortConflicts(t *testing.T) {
	db := newRepositoryTestDB(t, &model.Instance{}, &model.PortAllocation{})

	instanceID := int64(103)
	otherInstanceID := int64(104)
	instance := model.Instance{
		ID:          instanceID,
		UserID:      3,
		ChallengeID: 8,
		HostPort:    30004,
		Status:      model.InstanceStatusFailed,
		ShareScope:  model.InstanceSharingPerTeam,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30004, InstanceID: &otherInstanceID}).Error; err != nil {
		t.Fatalf("seed conflicting allocation: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30007, InstanceID: &instanceID}).Error; err != nil {
		t.Fatalf("seed rebound allocation: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	if err := repo.ResetInstanceRuntimeForRestart(context.Background(), instance.ID, model.InstanceStatusPending, time.Now().Add(2*time.Hour), true); err != nil {
		t.Fatalf("ResetInstanceRuntimeForRestart() error = %v", err)
	}

	var stored model.Instance
	if err := db.First(&stored, "id = ?", instance.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if stored.HostPort != 30007 {
		t.Fatalf("expected host port to switch to rebound allocation, got %d", stored.HostPort)
	}
}

func TestRepositoryIsHostPortReusableForRestart(t *testing.T) {
	db := newRepositoryTestDB(t, &model.PortAllocation{})

	currentInstanceID := int64(201)
	otherInstanceID := int64(202)
	if err := db.Create(&model.PortAllocation{Port: 30011, InstanceID: &currentInstanceID}).Error; err != nil {
		t.Fatalf("seed owned allocation: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30012, InstanceID: &otherInstanceID}).Error; err != nil {
		t.Fatalf("seed foreign allocation: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30013}).Error; err != nil {
		t.Fatalf("seed unbound allocation: %v", err)
	}

	repo := practiceinfra.NewRepository(db)

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
	if !reusable {
		t.Fatal("expected unbound host port to be reusable")
	}

	reusable, err = repo.IsHostPortReusableForRestart(context.Background(), currentInstanceID, 30014)
	if err != nil {
		t.Fatalf("missing IsHostPortReusableForRestart() error = %v", err)
	}
	if !reusable {
		t.Fatal("expected missing host port allocation to be reusable")
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
