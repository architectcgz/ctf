package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func TestUpdateStatusAndReleasePortWithContextSetsDestroyedAtForStoppedInstance(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)

	repo := NewRepository(db)
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	instance := model.Instance{
		ID:          1,
		UserID:      7,
		ChallengeID: 11,
		ContainerID: "inst-running",
		HostPort:    32001,
		Status:      model.InstanceStatusRunning,
		CreatedAt:   now.Add(-30 * time.Minute),
		UpdatedAt:   now.Add(-10 * time.Minute),
		ExpiresAt:   now.Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: instance.HostPort, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("seed port allocation: %v", err)
	}

	before := time.Now()
	if err := repo.UpdateStatusAndReleasePortWithContext(context.Background(), instance.ID, model.InstanceStatusStopped); err != nil {
		t.Fatalf("UpdateStatusAndReleasePortWithContext() error = %v", err)
	}
	after := time.Now()

	var row struct {
		Status      string     `gorm:"column:status"`
		DestroyedAt *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").Select("status", "destroyed_at").Where("id = ?", instance.ID).Take(&row).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if row.Status != model.InstanceStatusStopped {
		t.Fatalf("instance status = %q, want %q", row.Status, model.InstanceStatusStopped)
	}
	if row.DestroyedAt == nil {
		t.Fatal("expected destroyed_at to be set for stopped instance")
	}
	if row.DestroyedAt.Before(before.Add(-time.Second)) || row.DestroyedAt.After(after.Add(time.Second)) {
		t.Fatalf("destroyed_at = %v, want between %v and %v", row.DestroyedAt, before, after)
	}

	var remaining int64
	if err := db.Model(&model.PortAllocation{}).Where("instance_id = ? OR port = ?", instance.ID, instance.HostPort).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining port allocations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected port allocations to be released, got %d", remaining)
	}
}

func TestUpdateStatusAndReleasePortWithContextDoesNotSetDestroyedAtForFailedInstance(t *testing.T) {
	t.Parallel()

	db := newRuntimeRepositoryDestroyedAtTestDB(t)

	repo := NewRepository(db)
	instance := model.Instance{
		ID:          2,
		UserID:      9,
		ChallengeID: 15,
		ContainerID: "inst-creating",
		HostPort:    32002,
		Status:      model.InstanceStatusCreating,
		CreatedAt:   time.Now().Add(-5 * time.Minute),
		UpdatedAt:   time.Now().Add(-2 * time.Minute),
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}

	if err := repo.UpdateStatusAndReleasePortWithContext(context.Background(), instance.ID, model.InstanceStatusFailed); err != nil {
		t.Fatalf("UpdateStatusAndReleasePortWithContext() error = %v", err)
	}

	var row struct {
		Status      string     `gorm:"column:status"`
		DestroyedAt *time.Time `gorm:"column:destroyed_at"`
	}
	if err := db.Table("instances").Select("status", "destroyed_at").Where("id = ?", instance.ID).Take(&row).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if row.Status != model.InstanceStatusFailed {
		t.Fatalf("instance status = %q, want %q", row.Status, model.InstanceStatusFailed)
	}
	if row.DestroyedAt != nil {
		t.Fatalf("expected destroyed_at to stay nil for failed instance, got %v", row.DestroyedAt)
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
	if err := db.AutoMigrate(&model.Instance{}, &model.PortAllocation{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}
