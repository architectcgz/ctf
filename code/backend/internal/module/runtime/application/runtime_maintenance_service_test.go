package application

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

type maintenanceTestRepository struct {
	activeContainerIDs []string
}

func (r *maintenanceTestRepository) UpdateStatusAndReleasePort(int64, string) error {
	return nil
}

func (r *maintenanceTestRepository) FindExpired() ([]*model.Instance, error) {
	return nil, nil
}

func (r *maintenanceTestRepository) ListActiveContainerIDs() ([]string, error) {
	return append([]string(nil), r.activeContainerIDs...), nil
}

type maintenanceTestEngine struct {
	managedContainers []ManagedContainer
}

func (e *maintenanceTestEngine) ListManagedContainers(context.Context) ([]ManagedContainer, error) {
	return append([]ManagedContainer(nil), e.managedContainers...), nil
}

type maintenanceTestCleaner struct {
	removedContainerIDs []string
}

func (c *maintenanceTestCleaner) CleanupRuntimeWithContext(context.Context, *model.Instance) error {
	return nil
}

func (c *maintenanceTestCleaner) RemoveContainerWithContext(_ context.Context, containerID string) error {
	c.removedContainerIDs = append(c.removedContainerIDs, containerID)
	return nil
}

type typedNilMaintenanceEngine struct{}

func (*typedNilMaintenanceEngine) ListManagedContainers(context.Context) ([]ManagedContainer, error) {
	return nil, nil
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

	orphanContainers := selectOrphanContainers(managedContainers, activeContainerIDs, 5*time.Minute)
	if len(orphanContainers) != 1 {
		t.Fatalf("expected 1 orphan container, got %d (%v)", len(orphanContainers), orphanContainers)
	}
	if orphanContainers[0].ID != "orphan" {
		t.Fatalf("unexpected orphan container: %+v", orphanContainers[0])
	}
}

func TestRuntimeMaintenanceServiceCleanupOrphansSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		activeContainerIDs: []string{"active"},
	}
	engine := &maintenanceTestEngine{
		managedContainers: []ManagedContainer{
			{ID: "active", Name: "ctf-instance-active", CreatedAt: time.Now().Add(-10 * time.Minute)},
			{ID: "fresh", Name: "ctf-instance-fresh", CreatedAt: time.Now().Add(-2 * time.Minute)},
			{ID: "orphan", Name: "ctf-instance-orphan", CreatedAt: time.Now().Add(-12 * time.Minute)},
		},
	}
	cleaner := &maintenanceTestCleaner{}
	service := NewRuntimeMaintenanceService(repo, engine, cleaner, &config.ContainerConfig{
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	if err := service.CleanupOrphans(context.Background()); err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}
	if len(cleaner.removedContainerIDs) != 1 {
		t.Fatalf("expected 1 removed orphan container, got %v", cleaner.removedContainerIDs)
	}
	if cleaner.removedContainerIDs[0] != "orphan" {
		t.Fatalf("unexpected removed orphan container ids: %v", cleaner.removedContainerIDs)
	}
}

func TestNewRuntimeMaintenanceServiceTreatsTypedNilEngineAsNil(t *testing.T) {
	t.Parallel()

	var typedNil *typedNilMaintenanceEngine
	service := NewRuntimeMaintenanceService(&maintenanceTestRepository{}, typedNil, nil, &config.ContainerConfig{}, nil)
	if service.engine != nil {
		t.Fatalf("expected typed nil engine to be normalized to nil, got %#v", service.engine)
	}
}
