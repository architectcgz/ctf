package container

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/model"
)

func TestRepositoryListActiveContainerIDs(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 101,
		ContainerID: "running-container",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 102,
		ContainerID: "creating-container",
		Status:      model.InstanceStatusCreating,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 103,
		ContainerID: "stopped-container",
		Status:      model.InstanceStatusStopped,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 104,
		ContainerID: "",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})

	containerIDs, err := repo.ListActiveContainerIDs()
	if err != nil {
		t.Fatalf("ListActiveContainerIDs() error = %v", err)
	}
	if len(containerIDs) != 2 {
		t.Fatalf("expected 2 active container ids, got %d (%v)", len(containerIDs), containerIDs)
	}

	got := make(map[string]struct{}, len(containerIDs))
	for _, containerID := range containerIDs {
		got[containerID] = struct{}{}
	}
	if _, exists := got["running-container"]; !exists {
		t.Fatalf("running container not returned: %v", containerIDs)
	}
	if _, exists := got["creating-container"]; !exists {
		t.Fatalf("creating container not returned: %v", containerIDs)
	}
}

func TestSelectOrphanContainersSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	now := time.Now()
	managedContainers := []ManagedContainer{
		{ID: "active", Name: "ctf-instance-active", CreatedAt: now.Add(-10 * time.Minute)},
		{ID: "fresh", Name: "ctf-instance-fresh", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "orphan", Name: "ctf-instance-orphan", CreatedAt: now.Add(-12 * time.Minute)},
	}
	activeContainerIDs := map[string]struct{}{
		"active": {},
	}

	orphanContainers := selectOrphanContainers(managedContainers, activeContainerIDs, 5*time.Minute, now)
	if len(orphanContainers) != 1 {
		t.Fatalf("expected 1 orphan container, got %d (%v)", len(orphanContainers), orphanContainers)
	}
	if orphanContainers[0].ID != "orphan" {
		t.Fatalf("unexpected orphan container: %+v", orphanContainers[0])
	}
}

func TestManagedContainerLabels(t *testing.T) {
	t.Parallel()

	labels := managedContainerLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Instance{}); err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	return NewRepository(db)
}

func seedInstance(t *testing.T, db *gorm.DB, instance *model.Instance) {
	t.Helper()

	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
}
