package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type countRunningContextKey string

func newInstanceRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&instancecontracts.Instance{}); err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	return db
}

func TestCountRunningInstancesCountsOnlyRunningInstances(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)

	instances := []instancecontracts.Instance{
		{ID: 101, UserID: 9, ChallengeID: 21, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: time.Now().Add(time.Hour)},
		{ID: 102, UserID: 9, ChallengeID: 22, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: time.Now().Add(time.Hour)},
		{ID: 103, UserID: 9, ChallengeID: 23, Status: instancecontracts.InstanceStatusStopped, ExpiresAt: time.Now().Add(time.Hour)},
	}
	if err := db.Create(&instances).Error; err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	count, err := repo.CountRunningInstances(context.Background())
	if err != nil {
		t.Fatalf("CountRunningInstances() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("CountRunningInstances() count = %d, want 2", count)
	}
}

func TestCountRunningInstancesPropagatesContextToGORM(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)

	ctxKey := countRunningContextKey("count-running")
	expectedCtxValue := "ctx-count-running"
	callbackName := "count_running_instances_context_check"
	callbackCalled := false
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if got := tx.Statement.Context.Value(ctxKey); got == expectedCtxValue {
			callbackCalled = true
		}
	}); err != nil {
		t.Fatalf("register query callback: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Query().Remove(callbackName)
	})

	if err := db.Create(&instancecontracts.Instance{
		ID:          101,
		UserID:      9,
		ChallengeID: 21,
		Status:      instancecontracts.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("seed running instance: %v", err)
	}

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	count, err := repo.CountRunningInstances(ctx)
	if err != nil {
		t.Fatalf("CountRunningInstances() error = %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected gorm query callback to observe context-aware running instance count query")
	}
	if count != 1 {
		t.Fatalf("CountRunningInstances() count = %d, want 1", count)
	}
}
