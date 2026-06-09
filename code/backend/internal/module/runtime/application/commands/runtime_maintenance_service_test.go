package commands

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"ctf-platform/internal/config"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

type maintenanceTestRepository struct {
	activeContainerIDs                      []string
	recoverableActiveInstances              []*instanceentity.Instance
	stoppingInstances                       []*instanceentity.Instance
	runningWorkspaceByInstanceID            map[int64]*instanceports.AWDDefenseWorkspace
	requeuedIDs                             []int64
	finalizedStoppedIDs                     []int64
	operations                              []*instanceports.AWDServiceOperation
	finishedOperations                      []int64
	findExpiredFn                           func(ctx context.Context) ([]*instanceentity.Instance, error)
	listStoppingInstancesFn                 func(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error)
	listRecoverableActiveInstancesFn        func(ctx context.Context) ([]*instanceentity.Instance, error)
	findRunningWorkspaceByInstanceIDFn      func(ctx context.Context, instanceID int64) (*instanceports.AWDDefenseWorkspace, error)
	finalizeStoppedRuntimeFn                func(ctx context.Context, id int64) error
	requeueLostRuntimeFn                    func(ctx context.Context, id int64) (bool, error)
	listActiveContainerIDsFn                func(ctx context.Context) ([]string, error)
	updateStatusAndReleasePortFn            func(id int64, status string) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
}

type maintenanceTestLockStore struct {
	acquired bool
	called   int
}

func (s *maintenanceTestLockStore) WithStoppingCleanupLock(ctx context.Context, fn func(context.Context)) (bool, error) {
	s.called++
	if !s.acquired {
		return false, nil
	}
	fn(ctx)
	return true, nil
}

func (r *maintenanceTestRepository) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if r.updateStatusAndReleasePortWithContextFn != nil {
		return r.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return nil
}

func (r *maintenanceTestRepository) FindExpired(ctx context.Context) ([]*instanceentity.Instance, error) {
	if r.findExpiredFn != nil {
		return r.findExpiredFn(ctx)
	}
	return nil, nil
}

func (r *maintenanceTestRepository) ListStoppingInstances(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error) {
	if r.listStoppingInstancesFn != nil {
		return r.listStoppingInstancesFn(ctx, updatedBefore, limit)
	}
	return append([]*instanceentity.Instance(nil), r.stoppingInstances...), nil
}

func (r *maintenanceTestRepository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	if r.listActiveContainerIDsFn != nil {
		return r.listActiveContainerIDsFn(ctx)
	}
	return append([]string(nil), r.activeContainerIDs...), nil
}

func (r *maintenanceTestRepository) ListRecoverableActiveInstances(ctx context.Context) ([]*instanceentity.Instance, error) {
	if r.listRecoverableActiveInstancesFn != nil {
		return r.listRecoverableActiveInstancesFn(ctx)
	}
	return append([]*instanceentity.Instance(nil), r.recoverableActiveInstances...), nil
}

func (r *maintenanceTestRepository) FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*instanceports.AWDDefenseWorkspace, error) {
	if r.findRunningWorkspaceByInstanceIDFn != nil {
		return r.findRunningWorkspaceByInstanceIDFn(ctx, instanceID)
	}
	if r.runningWorkspaceByInstanceID == nil {
		return nil, nil
	}
	return r.runningWorkspaceByInstanceID[instanceID], nil
}

func (r *maintenanceTestRepository) CreateAWDServiceOperation(_ context.Context, operation *instanceports.AWDServiceOperation) error {
	operation.ID = int64(len(r.operations) + 1)
	r.operations = append(r.operations, operation)
	return nil
}

func (r *maintenanceTestRepository) FinishAWDServiceOperation(_ context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	r.finishedOperations = append(r.finishedOperations, operationID)
	for _, operation := range r.operations {
		if operation.ID == operationID {
			operation.Status = status
			operation.ErrorMessage = errorMessage
			operation.FinishedAt = &finishedAt
		}
	}
	return nil
}

func (r *maintenanceTestRepository) FinalizeStoppedRuntime(ctx context.Context, id int64) error {
	if r.finalizeStoppedRuntimeFn != nil {
		return r.finalizeStoppedRuntimeFn(ctx, id)
	}
	r.finalizedStoppedIDs = append(r.finalizedStoppedIDs, id)
	return nil
}

func (r *maintenanceTestRepository) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	if r.requeueLostRuntimeFn != nil {
		return r.requeueLostRuntimeFn(ctx, id)
	}
	r.requeuedIDs = append(r.requeuedIDs, id)
	return true, nil
}

type maintenanceTestEngine struct {
	managedContainers []instanceports.ManagedContainer
	containerStates   map[string]*instanceports.ManagedContainerState
	inspectErr        error
	inspectErrs       map[string]error
	startedIDs        []string
	startErrs         map[string]error
}

func (e *maintenanceTestEngine) ListManagedContainers(context.Context) ([]instanceports.ManagedContainer, error) {
	return append([]instanceports.ManagedContainer(nil), e.managedContainers...), nil
}

func (e *maintenanceTestEngine) InspectManagedContainer(_ context.Context, containerID string) (*instanceports.ManagedContainerState, error) {
	if e.inspectErr != nil {
		return nil, e.inspectErr
	}
	if err, ok := e.inspectErrs[containerID]; ok {
		return nil, err
	}
	if e.containerStates == nil {
		return &instanceports.ManagedContainerState{ID: containerID, Exists: true, Running: true, Status: "running"}, nil
	}
	if state, ok := e.containerStates[containerID]; ok {
		return state, nil
	}
	return &instanceports.ManagedContainerState{ID: containerID, Exists: false}, nil
}

func (e *maintenanceTestEngine) StartContainer(_ context.Context, containerID string) error {
	e.startedIDs = append(e.startedIDs, containerID)
	if err, ok := e.startErrs[containerID]; ok {
		return err
	}
	return nil
}

type maintenanceTestCleaner struct {
	removedContainerIDs []string
	cleanupInstanceIDs  []int64
	cleanupRuntimeFn    func(context.Context, *instanceentity.Instance) error
}

func (c *maintenanceTestCleaner) CleanupRuntime(ctx context.Context, instance *instanceentity.Instance) error {
	if instance != nil {
		c.cleanupInstanceIDs = append(c.cleanupInstanceIDs, instance.ID)
	}
	if c.cleanupRuntimeFn != nil {
		return c.cleanupRuntimeFn(ctx, instance)
	}
	return nil
}

func (c *maintenanceTestCleaner) RemoveContainer(_ context.Context, containerID string) error {
	c.removedContainerIDs = append(c.removedContainerIDs, containerID)
	return nil
}

type typedNilMaintenanceEngine struct{}

func (*typedNilMaintenanceEngine) ListManagedContainers(context.Context) ([]instanceports.ManagedContainer, error) {
	return nil, nil
}

func (*typedNilMaintenanceEngine) InspectManagedContainer(context.Context, string) (*instanceports.ManagedContainerState, error) {
	return nil, nil
}

func (*typedNilMaintenanceEngine) StartContainer(context.Context, string) error {
	return nil
}

func TestRuntimeMaintenanceServiceCleanupOrphansSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		activeContainerIDs: []string{"active"},
	}
	engine := &maintenanceTestEngine{
		managedContainers: []instanceports.ManagedContainer{
			{ID: "active", Name: "ctf-instance-active", CreatedAt: time.Now().Add(-10 * time.Minute)},
			{ID: "fresh", Name: "ctf-instance-fresh", CreatedAt: time.Now().Add(-2 * time.Minute)},
			{ID: "orphan", Name: "ctf-instance-orphan", CreatedAt: time.Now().Add(-12 * time.Minute)},
		},
	}
	cleaner := &maintenanceTestCleaner{}
	service := instancecmd.NewInstanceMaintenanceService(repo, engine, cleaner, &config.ContainerConfig{
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
	service := instancecmd.NewInstanceMaintenanceService(&maintenanceTestRepository{}, typedNil, nil, &config.ContainerConfig{}, nil)
	engineField := reflect.ValueOf(service).Elem().FieldByName("engine")
	if !engineField.IsNil() {
		t.Fatalf("expected typed nil engine to be normalized to nil, got %#v", engineField)
	}
}

type runtimeMaintenanceContextKey string

func TestRuntimeMaintenanceServiceCleanExpiredInstancesPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := runtimeMaintenanceContextKey("maintenance")
	expectedCtxValue := "ctx-runtime-maintenance"
	updateCalled := false
	repo := &maintenanceTestRepository{
		findExpiredFn: func(ctx context.Context) ([]*instanceentity.Instance, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-expired ctx value %v, got %v", expectedCtxValue, got)
			}
			return []*instanceentity.Instance{{ID: 41, HostPort: 30041}}, nil
		},
		updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update-status ctx value %v, got %v", expectedCtxValue, got)
			}
			if id != 41 || status != instanceentity.InstanceStatusExpired {
				t.Fatalf("unexpected update args: id=%d status=%s", id, status)
			}
			return nil
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, &maintenanceTestCleaner{}, &config.ContainerConfig{}, nil)

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
		managedContainers: []instanceports.ManagedContainer{
			{ID: "active", Name: "ctf-instance-active", CreatedAt: time.Now().Add(-10 * time.Minute)},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, engine, &maintenanceTestCleaner{}, &config.ContainerConfig{
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
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:          42,
				ContainerID: "missing-container",
				Status:      instanceentity.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	engine := &maintenanceTestEngine{
		containerStates: map[string]*instanceports.ManagedContainerState{
			"missing-container": {ID: "missing-container", Exists: false},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, engine, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 1 || repo.requeuedIDs[0] != 42 {
		t.Fatalf("expected instance 42 requeued, got %v", repo.requeuedIDs)
	}
}

func TestRuntimeMaintenanceServiceSkipsStoppingInstanceRecovery(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:          4201,
				ContainerID: "stopping-container",
				Status:      instanceentity.InstanceStatusStopping,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	engine := &maintenanceTestEngine{
		containerStates: map[string]*instanceports.ManagedContainerState{
			"stopping-container": {ID: "stopping-container", Exists: false},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, engine, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(repo.requeuedIDs) != 0 {
		t.Fatalf("expected stopping instance to be skipped, got requeue ids %v", repo.requeuedIDs)
	}
}

func TestRuntimeMaintenanceServiceRunStoppingCleanupLoopFinalizesStoppingInstances(t *testing.T) {
	t.Parallel()

	var repoMu sync.Mutex
	repo := &maintenanceTestRepository{
		stoppingInstances: []*instanceentity.Instance{
			{
				ID:          4301,
				ContainerID: "stopping-container",
				Status:      instanceentity.InstanceStatusStopping,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	repo.listStoppingInstancesFn = func(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error) {
		repoMu.Lock()
		defer repoMu.Unlock()
		return append([]*instanceentity.Instance(nil), repo.stoppingInstances...), nil
	}
	repo.finalizeStoppedRuntimeFn = func(ctx context.Context, id int64) error {
		repoMu.Lock()
		defer repoMu.Unlock()
		repo.finalizedStoppedIDs = append(repo.finalizedStoppedIDs, id)
		filtered := repo.stoppingInstances[:0]
		for _, item := range repo.stoppingInstances {
			if item == nil || item.ID == id {
				continue
			}
			filtered = append(filtered, item)
		}
		repo.stoppingInstances = filtered
		return nil
	}
	cleaner := &maintenanceTestCleaner{}
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, cleaner, &config.ContainerConfig{
		DeletePollInterval:  5 * time.Millisecond,
		DeleteMaxConcurrent: 2,
	}, nil)

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunStoppingCleanupLoop(runCtx)
	}()

	deadline := time.Now().Add(time.Second)
	for len(repo.finalizedStoppedIDs) < 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done

	if len(cleaner.cleanupInstanceIDs) != 1 || cleaner.cleanupInstanceIDs[0] != 4301 {
		t.Fatalf("expected stopping instance cleanup, got %v", cleaner.cleanupInstanceIDs)
	}
	if len(repo.finalizedStoppedIDs) != 1 || repo.finalizedStoppedIDs[0] != 4301 {
		t.Fatalf("expected stopping instance finalize, got %v", repo.finalizedStoppedIDs)
	}
}

func TestRuntimeMaintenanceServiceRunStoppingCleanupLoopSkipsWhenLockHeldByAnotherNode(t *testing.T) {
	t.Parallel()

	var queried bool
	repo := &maintenanceTestRepository{}
	repo.listStoppingInstancesFn = func(context.Context, time.Time, int) ([]*instanceentity.Instance, error) {
		queried = true
		return nil, nil
	}
	lockStore := &maintenanceTestLockStore{acquired: false}
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, &maintenanceTestCleaner{}, &config.ContainerConfig{
		DeletePollInterval:  5 * time.Millisecond,
		DeleteMaxConcurrent: 2,
	}, nil, lockStore)

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunStoppingCleanupLoop(runCtx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if lockStore.called == 0 {
		t.Fatal("expected stopping cleanup lock to be attempted")
	}
	if queried {
		t.Fatal("expected stopping instances query to be skipped when lock is not acquired")
	}
}

func TestRuntimeMaintenanceServiceRunStoppingCleanupLoopPassesConfiguredBatchLimit(t *testing.T) {
	t.Parallel()

	limits := make(chan int, 1)
	repo := &maintenanceTestRepository{}
	repo.listStoppingInstancesFn = func(_ context.Context, _ time.Time, limit int) ([]*instanceentity.Instance, error) {
		limits <- limit
		return nil, nil
	}
	lockStore := &maintenanceTestLockStore{acquired: true}
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, &maintenanceTestCleaner{}, &config.ContainerConfig{
		DeletePollInterval:  5 * time.Millisecond,
		DeleteMaxConcurrent: 3,
	}, nil, lockStore)

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunStoppingCleanupLoop(runCtx)
	}()

	var gotLimit int
	select {
	case gotLimit = <-limits:
	case <-time.After(time.Second):
		t.Fatal("expected stopping instances query")
	}
	cancel()
	<-done

	if gotLimit != 3 {
		t.Fatalf("expected configured batch limit 3, got %d", gotLimit)
	}
	if lockStore.called == 0 {
		t.Fatal("expected stopping cleanup lock to be attempted")
	}
}

func TestRuntimeMaintenanceServiceStoppingCleanupWakeupTriggersDispatchBeforeNextTick(t *testing.T) {
	t.Parallel()

	firstQueryDone := make(chan struct{})
	var (
		repoMu sync.Mutex
		calls  int
	)
	repo := &maintenanceTestRepository{
		stoppingInstances: []*instanceentity.Instance{
			{ID: 4501, Status: instanceentity.InstanceStatusStopping, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	repo.listStoppingInstancesFn = func(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error) {
		repoMu.Lock()
		defer repoMu.Unlock()
		calls++
		if calls == 1 {
			close(firstQueryDone)
			return nil, nil
		}
		return append([]*instanceentity.Instance(nil), repo.stoppingInstances...), nil
	}
	repo.finalizeStoppedRuntimeFn = func(ctx context.Context, id int64) error {
		repoMu.Lock()
		defer repoMu.Unlock()
		repo.finalizedStoppedIDs = append(repo.finalizedStoppedIDs, id)
		repo.stoppingInstances = nil
		return nil
	}

	bus := platformevents.NewBus()
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, &maintenanceTestCleaner{}, &config.ContainerConfig{
		DeletePollInterval:  time.Hour,
		DeleteMaxConcurrent: 2,
	}, nil)
	service.RegisterStoppingCleanupWakeup(bus)

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunStoppingCleanupLoop(runCtx)
	}()

	select {
	case <-firstQueryDone:
	case <-time.After(time.Second):
		t.Fatal("expected initial stopping cleanup query")
	}
	if err := bus.Publish(context.Background(), platformevents.Event{
		Name: instancecontracts.EventInstanceStoppingCleanupWakeup,
		Payload: instancecontracts.InstanceStoppingCleanupWakeupEvent{
			InstanceID: 4501,
		},
	}); err != nil {
		t.Fatalf("publish wakeup event: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for len(repo.finalizedStoppedIDs) < 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done

	if len(repo.finalizedStoppedIDs) != 1 || repo.finalizedStoppedIDs[0] != 4501 {
		t.Fatalf("expected wakeup to finalize stopping instance before next tick, got %v", repo.finalizedStoppedIDs)
	}
}

func TestRuntimeMaintenanceServiceRunStoppingCleanupLoopHonorsConcurrencyLimit(t *testing.T) {
	t.Parallel()

	var repoMu sync.Mutex
	repo := &maintenanceTestRepository{
		stoppingInstances: []*instanceentity.Instance{
			{ID: 4401, Status: instanceentity.InstanceStatusStopping, ExpiresAt: time.Now().Add(time.Hour)},
			{ID: 4402, Status: instanceentity.InstanceStatusStopping, ExpiresAt: time.Now().Add(time.Hour)},
			{ID: 4403, Status: instanceentity.InstanceStatusStopping, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	repo.listStoppingInstancesFn = func(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error) {
		repoMu.Lock()
		defer repoMu.Unlock()
		return append([]*instanceentity.Instance(nil), repo.stoppingInstances...), nil
	}
	repo.finalizeStoppedRuntimeFn = func(ctx context.Context, id int64) error {
		repoMu.Lock()
		defer repoMu.Unlock()
		repo.finalizedStoppedIDs = append(repo.finalizedStoppedIDs, id)
		filtered := repo.stoppingInstances[:0]
		for _, item := range repo.stoppingInstances {
			if item == nil || item.ID == id {
				continue
			}
			filtered = append(filtered, item)
		}
		repo.stoppingInstances = filtered
		return nil
	}
	started := make(chan int64, 3)
	release := make(chan struct{})
	var (
		mu            sync.Mutex
		active        int
		maxConcurrent int
	)
	cleaner := &maintenanceTestCleaner{
		cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			mu.Lock()
			active++
			if active > maxConcurrent {
				maxConcurrent = active
			}
			mu.Unlock()

			started <- instance.ID

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-release:
			}

			mu.Lock()
			active--
			mu.Unlock()
			return nil
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, cleaner, &config.ContainerConfig{
		DeletePollInterval:  5 * time.Millisecond,
		DeleteMaxConcurrent: 2,
	}, nil)

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunStoppingCleanupLoop(runCtx)
	}()

	first := <-started
	second := <-started
	if first == second {
		t.Fatalf("expected distinct stopping instances to start cleanup, got %d twice", first)
	}

	select {
	case third := <-started:
		t.Fatalf("expected cleanup concurrency to stay capped at 2, got third start for %d before release", third)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	deadline := time.Now().Add(time.Second)
	for len(repo.finalizedStoppedIDs) < 3 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done

	if len(repo.finalizedStoppedIDs) != 3 {
		t.Fatalf("expected all stopping instances finalized, got %v", repo.finalizedStoppedIDs)
	}
	if maxConcurrent != 2 {
		t.Fatalf("expected max concurrent cleanup workers = 2, got %d", maxConcurrent)
	}
}

func TestRuntimeMaintenanceServiceRestartsExitedTopologyContainerBeforeRequeue(t *testing.T) {
	t.Parallel()

	contestID := int64(9001)
	teamID := int64(9101)
	serviceID := int64(9201)
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{ContainerID: "entry", IsEntryPoint: true},
			{ContainerID: "sidecar"},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}
	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:             43,
				ContestID:      &contestID,
				TeamID:         &teamID,
				ServiceID:      &serviceID,
				ContainerID:    "entry",
				RuntimeDetails: runtimeDetails,
				Status:         instanceentity.InstanceStatusRunning,
				ExpiresAt:      time.Now().Add(time.Hour),
				UpdatedAt:      time.Now().Add(-time.Minute),
			},
		},
	}
	engine := &maintenanceTestEngine{
		containerStates: map[string]*instanceports.ManagedContainerState{
			"entry":   {ID: "entry", Exists: true, Running: true, Status: "running"},
			"sidecar": {ID: "sidecar", Exists: true, Running: false, Status: "exited"},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, engine, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(engine.startedIDs) != 1 || engine.startedIDs[0] != "sidecar" {
		t.Fatalf("expected stopped sidecar to be started first, got %v", engine.startedIDs)
	}
	if len(repo.requeuedIDs) != 0 {
		t.Fatalf("expected no requeue when stopped container starts, got %v", repo.requeuedIDs)
	}
	if len(repo.operations) != 1 {
		t.Fatalf("expected one system recover operation, got %+v", repo.operations)
	}
	operation := repo.operations[0]
	if operation.OperationType != instanceports.AWDServiceOperationTypeRecover || operation.RequestedBy != instanceports.AWDServiceOperationRequestedBySystem || operation.SLABillable {
		t.Fatalf("unexpected recover operation: %+v", operation)
	}
	if operation.Status != instanceports.AWDServiceOperationStatusRecovered || operation.FinishedAt == nil {
		t.Fatalf("expected recovered operation to be finished, got %+v", operation)
	}
}

func TestRuntimeMaintenanceServiceRestartsStoppedWorkspaceCompanionBeforeRequeue(t *testing.T) {
	t.Parallel()

	contestID := int64(9301)
	teamID := int64(9401)
	serviceID := int64(9501)
	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:          48,
				ContestID:   &contestID,
				TeamID:      &teamID,
				ServiceID:   &serviceID,
				ContainerID: "entry",
				Status:      instanceentity.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
		runningWorkspaceByInstanceID: map[int64]*instanceports.AWDDefenseWorkspace{
			48: {
				ContainerID: "workspace-companion",
			},
		},
	}
	engine := &maintenanceTestEngine{
		containerStates: map[string]*instanceports.ManagedContainerState{
			"entry":               {ID: "entry", Exists: true, Running: true, Status: "running"},
			"workspace-companion": {ID: "workspace-companion", Exists: true, Running: false, Status: "exited"},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, engine, nil, &config.ContainerConfig{
		CreateTimeout: 30 * time.Second,
	}, nil)

	if err := service.ReconcileLostActiveRuntimes(context.Background()); err != nil {
		t.Fatalf("ReconcileLostActiveRuntimes() error = %v", err)
	}
	if len(engine.startedIDs) != 1 || engine.startedIDs[0] != "workspace-companion" {
		t.Fatalf("expected stopped workspace companion to be started first, got %v", engine.startedIDs)
	}
	if len(repo.requeuedIDs) != 0 {
		t.Fatalf("expected no requeue when workspace companion starts, got %v", repo.requeuedIDs)
	}
	if len(repo.operations) != 1 {
		t.Fatalf("expected one system recover operation, got %+v", repo.operations)
	}
}

func TestRuntimeMaintenanceServiceSkipsFreshCreatingInstanceWithoutContainer(t *testing.T) {
	t.Parallel()

	repo := &maintenanceTestRepository{
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:        44,
				Status:    instanceentity.InstanceStatusCreating,
				ExpiresAt: time.Now().Add(time.Hour),
				UpdatedAt: time.Now(),
			},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, &maintenanceTestEngine{}, nil, &config.ContainerConfig{
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
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:          45,
				ContainerID: "runtime",
				Status:      instanceentity.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, &maintenanceTestEngine{
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
		recoverableActiveInstances: []*instanceentity.Instance{
			{
				ID:          46,
				ContainerID: "inspect-fails",
				Status:      instanceentity.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
			{
				ID:          47,
				ContainerID: "missing-runtime",
				Status:      instanceentity.InstanceStatusRunning,
				ExpiresAt:   time.Now().Add(time.Hour),
				UpdatedAt:   time.Now().Add(-time.Minute),
			},
		},
	}
	service := instancecmd.NewInstanceMaintenanceService(repo, &maintenanceTestEngine{
		inspectErrs: map[string]error{
			"inspect-fails": fmt.Errorf("docker inspect failed"),
		},
		containerStates: map[string]*instanceports.ManagedContainerState{
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
