package runtime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

func TestServiceCleanupRuntimeFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	cleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil, nil)
	instance := &instanceentity.Instance{
		ID:          3002,
		ContainerID: "ctr-missing-engine",
		NetworkID:   "net-missing-engine",
	}

	err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(instance))
	if err == nil {
		t.Fatal("expected CleanupRuntime() to fail when runtime engine is unavailable")
	}
	if !errors.Is(err, runtimeports.ErrRuntimeEngineUnavailable) {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
}

func TestServiceCleanupRuntimeHonorsCancellation(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			if containerID != "ctr-3001" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	instance := &instanceentity.Instance{
		ID:          3001,
		ContainerID: "ctr-3001",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cleanupService.CleanupRuntime(ctx, runtimeCleanupTarget(instance)); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceCleanupRuntimeIgnoresMissingNetwork(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeNetworkErr: runtimeports.WrapRuntimeNetworkNotFound(errors.New("Error response from daemon: network net-missing not found")),
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	instance := &instanceentity.Instance{
		ID:        3004,
		NetworkID: "net-missing",
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(instance)); err != nil {
		t.Fatalf("expected missing network to be ignored, got %v", err)
	}
	if engine.removedNetworkID != "net-missing" {
		t.Fatalf("expected cleanup to attempt removing net-missing, got %s", engine.removedNetworkID)
	}
}

func TestServiceCleanupRuntimeRetriesDeadlineExceededNetworkRemoval(t *testing.T) {
	t.Parallel()

	removeCalls := 0
	engine := &fakeRuntimeEngine{
		removeNetworkFn: func(ctx context.Context, networkID string) error {
			removeCalls++
			if networkID != "net-timeout" {
				t.Fatalf("unexpected network id: %s", networkID)
			}
			if removeCalls == 1 {
				return context.DeadlineExceeded
			}
			return runtimeports.WrapRuntimeNetworkNotFound(errors.New("Error response from daemon: network net-timeout not found"))
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	instance := &instanceentity.Instance{
		ID:        30041,
		NetworkID: "net-timeout",
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(instance)); err != nil {
		t.Fatalf("expected network timeout followed by not found to be treated as success, got %v", err)
	}
	if removeCalls != 2 {
		t.Fatalf("expected cleanup to retry remove network once after deadline exceeded, got %d calls", removeCalls)
	}
}

func TestServiceCleanupRuntimeRetriesDeadlineExceededContainerRemoval(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instanceID := int64(3005)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30009,
		InstanceID: &instanceID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	removeCalls := 0
	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			removeCalls++
			if containerID != "ctr-deadline" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			if removeCalls == 1 {
				return context.DeadlineExceeded
			}
			return runtimeports.WrapRuntimeContainerNotFound(errors.New("Error response from daemon: No such container: ctr-deadline"))
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)

	instance := &instanceentity.Instance{
		ID:          instanceID,
		UserID:      1,
		ChallengeID: 101,
		ContainerID: "ctr-deadline",
		HostPort:    30009,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(instance)); err != nil {
		t.Fatalf("expected timeout followed by missing container to be treated as success, got %v", err)
	}
	if removeCalls != 2 {
		t.Fatalf("expected cleanup to retry remove once after deadline exceeded, got %d calls", removeCalls)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30009).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be released, count=%d", count)
	}
}

func TestServiceCleanupRuntimeWaitsForContainerRemovalAlreadyInProgress(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instanceID := int64(3006)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30010,
		InstanceID: &instanceID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	removeCalls := 0
	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			removeCalls++
			if containerID != "ctr-removing" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			switch removeCalls {
			case 1:
				return context.DeadlineExceeded
			case 2:
				return errors.New("Error response from daemon: removal of container ctr-removing is already in progress")
			default:
				return runtimeports.WrapRuntimeContainerNotFound(errors.New("Error response from daemon: No such container: ctr-removing"))
			}
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)

	instance := &instanceentity.Instance{
		ID:          instanceID,
		UserID:      2,
		ChallengeID: 102,
		ContainerID: "ctr-removing",
		HostPort:    30010,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := cleanupService.CleanupRuntime(context.Background(), runtimeCleanupTarget(instance)); err != nil {
		t.Fatalf("expected cleanup to wait until background removal completes, got %v", err)
	}
	if removeCalls != 3 {
		t.Fatalf("expected cleanup to poll until container removal finished, got %d calls", removeCalls)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30010).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be released, count=%d", count)
	}
}
