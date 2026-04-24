package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type maintenanceTestRepository struct {
	activeContainerIDs                      []string
	findExpiredFn                           func() ([]*model.Instance, error)
	updateStatusAndReleasePortFn            func(id int64, status string) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
}

func (r *maintenanceTestRepository) UpdateStatusAndReleasePort(int64, string) error {
	if r.updateStatusAndReleasePortFn != nil {
		return r.updateStatusAndReleasePortFn(int64(0), "")
	}
	return nil
}

func (r *maintenanceTestRepository) UpdateStatusAndReleasePortWithContext(ctx context.Context, id int64, status string) error {
	if r.updateStatusAndReleasePortWithContextFn != nil {
		return r.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return nil
}

func (r *maintenanceTestRepository) FindExpired() ([]*model.Instance, error) {
	if r.findExpiredFn != nil {
		return r.findExpiredFn()
	}
	return nil, nil
}

func (r *maintenanceTestRepository) ListActiveContainerIDs() ([]string, error) {
	return append([]string(nil), r.activeContainerIDs...), nil
}

type maintenanceTestEngine struct {
	managedContainers []runtimeports.ManagedContainer
}

func (e *maintenanceTestEngine) ListManagedContainers(context.Context) ([]runtimeports.ManagedContainer, error) {
	return append([]runtimeports.ManagedContainer(nil), e.managedContainers...), nil
}

type maintenanceTestCleaner struct {
	removedContainerIDs []string
}

func (c *maintenanceTestCleaner) CleanupRuntime(context.Context, *model.Instance) error {
	return nil
}

func (c *maintenanceTestCleaner) RemoveContainer(_ context.Context, containerID string) error {
	c.removedContainerIDs = append(c.removedContainerIDs, containerID)
	return nil
}

type typedNilMaintenanceEngine struct{}

func (*typedNilMaintenanceEngine) ListManagedContainers(context.Context) ([]runtimeports.ManagedContainer, error) {
	return nil, nil
}

func TestSelectOrphanContainersSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	now := time.Now()
	managedContainers := []runtimeports.ManagedContainer{
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
		managedContainers: []runtimeports.ManagedContainer{
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

type runtimeMaintenanceContextKey string

func TestRuntimeMaintenanceServiceCleanExpiredInstancesPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := runtimeMaintenanceContextKey("maintenance")
	expectedCtxValue := "ctx-runtime-maintenance"
	updateCalled := false
	repo := &maintenanceTestRepository{
		findExpiredFn: func() ([]*model.Instance, error) {
			return []*model.Instance{{ID: 41, HostPort: 30041}}, nil
		},
		updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update-status ctx value %v, got %v", expectedCtxValue, got)
			}
			if id != 41 || status != model.InstanceStatusExpired {
				t.Fatalf("unexpected update args: id=%d status=%s", id, status)
			}
			return nil
		},
	}
	service := NewRuntimeMaintenanceService(repo, nil, &maintenanceTestCleaner{}, &config.ContainerConfig{}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.CleanExpiredInstances(ctx); err != nil {
		t.Fatalf("CleanExpiredInstances() error = %v", err)
	}
	if !updateCalled {
		t.Fatal("expected update status repository to be called")
	}
}
