package commands

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type maintenanceTestRepository struct {
	activeContainerIDs                      []string
	recoverableActiveInstances              []*model.Instance
	requeuedIDs                             []int64
	findExpiredFn                           func(ctx context.Context) ([]*model.Instance, error)
	listRecoverableActiveInstancesFn        func(ctx context.Context) ([]*model.Instance, error)
	requeueLostRuntimeFn                    func(ctx context.Context, id int64) (bool, error)
	listActiveContainerIDsFn                func(ctx context.Context) ([]string, error)
	updateStatusAndReleasePortFn            func(id int64, status string) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
}

func (r *maintenanceTestRepository) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if r.updateStatusAndReleasePortWithContextFn != nil {
		return r.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return nil
}

func (r *maintenanceTestRepository) FindExpired(ctx context.Context) ([]*model.Instance, error) {
	if r.findExpiredFn != nil {
		return r.findExpiredFn(ctx)
	}
	return nil, nil
}

func (r *maintenanceTestRepository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	if r.listActiveContainerIDsFn != nil {
		return r.listActiveContainerIDsFn(ctx)
	}
	return append([]string(nil), r.activeContainerIDs...), nil
}

func (r *maintenanceTestRepository) ListRecoverableActiveInstances(ctx context.Context) ([]*model.Instance, error) {
	if r.listRecoverableActiveInstancesFn != nil {
		return r.listRecoverableActiveInstancesFn(ctx)
	}
	return append([]*model.Instance(nil), r.recoverableActiveInstances...), nil
}

func (r *maintenanceTestRepository) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	if r.requeueLostRuntimeFn != nil {
		return r.requeueLostRuntimeFn(ctx, id)
	}
	r.requeuedIDs = append(r.requeuedIDs, id)
	return true, nil
}

type maintenanceTestEngine struct {
	managedContainers []runtimeports.ManagedContainer
	containerStates   map[string]*runtimeports.ManagedContainerState
	inspectErr        error
	inspectErrs       map[string]error
}

func (e *maintenanceTestEngine) ListManagedContainers(context.Context) ([]runtimeports.ManagedContainer, error) {
	return append([]runtimeports.ManagedContainer(nil), e.managedContainers...), nil
}

func (e *maintenanceTestEngine) InspectManagedContainer(_ context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if e.inspectErr != nil {
		return nil, e.inspectErr
	}
	if err, ok := e.inspectErrs[containerID]; ok {
		return nil, err
	}
	if e.containerStates == nil {
		return &runtimeports.ManagedContainerState{ID: containerID, Exists: true, Running: true, Status: "running"}, nil
	}
	if state, ok := e.containerStates[containerID]; ok {
		return state, nil
	}
	return &runtimeports.ManagedContainerState{ID: containerID, Exists: false}, nil
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

func (*typedNilMaintenanceEngine) InspectManagedContainer(context.Context, string) (*runtimeports.ManagedContainerState, error) {
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
		findExpiredFn: func(ctx context.Context) ([]*model.Instance, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-expired ctx value %v, got %v", expectedCtxValue, got)
			}
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

func TestRuntimeMaintenanceServiceCleanupOrphansPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := runtimeMaintenanceContextKey("orphan-maintenance")
	expectedCtxValue := "ctx-orphan-maintenance"
	repo := &maintenanceTestRepository{
		listActiveContainerIDsFn: func(ctx context.Context) ([]string, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-active ctx value %v, got %v", expectedCtxValue, got)
			}
			return []string{"active"}, nil
		},
	}
	engine := &maintenanceTestEngine{
		managedContainers: []runtimeports.ManagedContainer{
			{ID: "active", Name: "ctf-instance-active", CreatedAt: time.Now().Add(-10 * time.Minute)},
		},
	}
	service := NewRuntimeMaintenanceService(repo, engine, &maintenanceTestCleaner{}, &config.ContainerConfig{
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.CleanupOrphans(ctx); err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}
}

func TestRuntimeMaintenanceServiceRequeuesMissingRunningContainer(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*model.Instance{
			{
				ID:          42,
				ContainerID: "missing-container",
				Status:      model.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	engine := &maintenanceTestEngine{
		containerStates: map[string]*runtimeports.ManagedContainerState{
			"missing-container": {ID: "missing-container", Exists: false},
		},
	}
	service := NewRuntimeMaintenanceService(repo, engine, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 1 || repo.requeuedIDs[0] != 42 {
		t.Fatalf("expected instance 42 requeued, got %v", repo.requeuedIDs)
	}
}

func TestRuntimeMaintenanceServiceRequeuesExitedTopologyContainer(t *testing.T) {
	t.Parallel()

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Containers: []model.InstanceRuntimeContainer{
			{ContainerID: "entry", IsEntryPoint: true},
			{ContainerID: "sidecar"},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}
	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*model.Instance{
			{
				ID:             43,
				ContainerID:    "entry",
				RuntimeDetails: runtimeDetails,
				Status:         model.InstanceStatusRunning,
				ExpiresAt:      time.Now().Add(time.Hour),
				UpdatedAt:      time.Now().Add(-time.Minute),
			},
		},
	}
	engine := &maintenanceTestEngine{
		containerStates: map[string]*runtimeports.ManagedContainerState{
			"entry":   {ID: "entry", Exists: true, Running: true, Status: "running"},
			"sidecar": {ID: "sidecar", Exists: true, Running: false, Status: "exited"},
		},
	}
	service := NewRuntimeMaintenanceService(repo, engine, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 1 || repo.requeuedIDs[0] != 43 {
		t.Fatalf("expected instance 43 requeued, got %v", repo.requeuedIDs)
	}
}

func TestRuntimeMaintenanceServiceSkipsFreshCreatingInstanceWithoutContainer(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*model.Instance{
			{
				ID:        44,
				Status:    model.InstanceStatusCreating,
				ExpiresAt: time.Now().Add(time.Hour),
				UpdatedAt: time.Now(),
			},
		},
	}
	service := NewRuntimeMaintenanceService(repo, &maintenanceTestEngine{}, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 0 {
		t.Fatalf("expected fresh creating instance not requeued, got %v", repo.requeuedIDs)
	}
}

func TestRuntimeMaintenanceServiceSkipsInstanceWhenDockerInspectFails(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*model.Instance{
			{
				ID:          45,
				ContainerID: "runtime",
				Status:      model.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	service := NewRuntimeMaintenanceService(repo, &maintenanceTestEngine{
		inspectErr: fmt.Errorf("docker unavailable"),
	}, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 0 {
		t.Fatalf("expected no requeue on docker inspect error, got %v", repo.requeuedIDs)
	}
}

func TestRuntimeMaintenanceServiceInspectFailureDoesNotBlockOtherInstances(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*model.Instance{
			{
				ID:          46,
				ContainerID: "inspect-fails",
				Status:      model.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
			{
				ID:          47,
				ContainerID: "missing-runtime",
				Status:      model.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	service := NewRuntimeMaintenanceService(repo, &maintenanceTestEngine{
		inspectErrs: map[string]error{
			"inspect-fails": fmt.Errorf("docker inspect failed"),
		},
		containerStates: map[string]*runtimeports.ManagedContainerState{
			"missing-runtime": {ID: "missing-runtime", Exists: false},
		},
	}, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 1 || repo.requeuedIDs[0] != 47 {
		t.Fatalf("expected only instance 47 requeued, got %v", repo.requeuedIDs)
	}
}
