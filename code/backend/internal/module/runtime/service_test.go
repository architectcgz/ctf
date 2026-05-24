package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/internal/shared/taxonomy"
)

func TestRepositoryListActiveContainerIDs(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	if err := repo.db.AutoMigrate(&runtimeentity.AWDDefenseWorkspace{}); err != nil {
		t.Fatalf("migrate awd defense workspace table: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		UserID:      1,
		ChallengeID: 101,
		ContainerID: "running-container",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &instanceentity.Instance{
		UserID:         1,
		ChallengeID:    102,
		ContainerID:    "creating-container",
		RuntimeDetails: `{"containers":[{"container_id":"sidecar-1"},{"container_id":"creating-container"}]}`,
		Status:         instanceentity.InstanceStatusCreating,
		ExpiresAt:      time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &instanceentity.Instance{
		UserID:      1,
		ChallengeID: 103,
		ContainerID: "stopped-container",
		Status:      instanceentity.InstanceStatusStopped,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &instanceentity.Instance{
		UserID:      1,
		ChallengeID: 104,
		ContainerID: "",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	contestID := int64(301)
	teamID := int64(401)
	serviceID := int64(501)
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1005,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 105,
		ContainerID: "runtime-awd",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedAWDDefenseWorkspace(t, repo.db, &runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        1005,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-running",
		SeedSignature:     "seed-running",
	})
	stoppedContestID := int64(302)
	stoppedTeamID := int64(402)
	stoppedServiceID := int64(502)
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1006,
		UserID:      1,
		ContestID:   &stoppedContestID,
		TeamID:      &stoppedTeamID,
		ServiceID:   &stoppedServiceID,
		ChallengeID: 106,
		ContainerID: "runtime-awd-stopped",
		Status:      instanceentity.InstanceStatusStopped,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedAWDDefenseWorkspace(t, repo.db, &runtimeentity.AWDDefenseWorkspace{
		ContestID:         stoppedContestID,
		TeamID:            stoppedTeamID,
		ServiceID:         stoppedServiceID,
		InstanceID:        1006,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-stopped",
		SeedSignature:     "seed-stopped",
	})
	stoppingContestID := int64(304)
	stoppingTeamID := int64(404)
	stoppingServiceID := int64(504)
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1008,
		UserID:      1,
		ContestID:   &stoppingContestID,
		TeamID:      &stoppingTeamID,
		ServiceID:   &stoppingServiceID,
		ChallengeID: 108,
		ContainerID: "runtime-awd-stopping",
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedAWDDefenseWorkspace(t, repo.db, &runtimeentity.AWDDefenseWorkspace{
		ContestID:         stoppingContestID,
		TeamID:            stoppingTeamID,
		ServiceID:         stoppingServiceID,
		InstanceID:        1008,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-stopping",
		SeedSignature:     "seed-stopping",
	})
	failedContestID := int64(303)
	failedTeamID := int64(403)
	failedServiceID := int64(503)
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1007,
		UserID:      1,
		ContestID:   &failedContestID,
		TeamID:      &failedTeamID,
		ServiceID:   &failedServiceID,
		ChallengeID: 107,
		ContainerID: "runtime-awd-failed",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedAWDDefenseWorkspace(t, repo.db, &runtimeentity.AWDDefenseWorkspace{
		ContestID:         failedContestID,
		TeamID:            failedTeamID,
		ServiceID:         failedServiceID,
		InstanceID:        1007,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusFailed,
		ContainerID:       "workspace-failed",
		SeedSignature:     "seed-failed",
	})

	containerIDs, err := repo.ListActiveContainerIDs(context.Background())
	if err != nil {
		t.Fatalf("ListActiveContainerIDs() error = %v", err)
	}
	if len(containerIDs) != 8 {
		t.Fatalf("expected 8 active container ids, got %d (%v)", len(containerIDs), containerIDs)
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
	if _, exists := got["sidecar-1"]; !exists {
		t.Fatalf("sidecar container not returned: %v", containerIDs)
	}
	if _, exists := got["workspace-running"]; !exists {
		t.Fatalf("running workspace container not returned: %v", containerIDs)
	}
	if _, exists := got["runtime-awd-stopping"]; !exists {
		t.Fatalf("stopping runtime container not returned: %v", containerIDs)
	}
	if _, exists := got["workspace-stopping"]; !exists {
		t.Fatalf("stopping workspace container not returned: %v", containerIDs)
	}
	if _, exists := got["workspace-stopped"]; exists {
		t.Fatalf("workspace container for stopped instance should not be returned: %v", containerIDs)
	}
	if _, exists := got["workspace-failed"]; exists {
		t.Fatalf("failed workspace container should not be returned: %v", containerIDs)
	}
}

func TestRepositoryUpdateStatusAndReleasePortRemovesAllocation(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instance := &instanceentity.Instance{
		ID:          2001,
		UserID:      1,
		ChallengeID: 101,
		HostPort:    30001,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30001,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}
	if err := repo.db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.0.0/24",
		InstanceID: &instance.ID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create network allocation: %v", err)
	}

	if err := repo.UpdateStatusAndReleasePort(context.Background(), instance.ID, instanceentity.InstanceStatusFailed); err != nil {
		t.Fatalf("UpdateStatusAndReleasePort() error = %v", err)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != instanceentity.InstanceStatusFailed {
		t.Fatalf("expected failed status, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30001).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
	if err := repo.db.Model(&runtimeentity.NetworkAllocation{}).Where("instance_id = ?", instance.ID).Count(&count).Error; err != nil {
		t.Fatalf("count network allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected network allocation to be removed, count=%d", count)
	}
}

func TestRepositoryListStoppingInstancesFiltersByUpdatedBefore(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	cutoff := now.Add(-5 * time.Minute)
	staleUpdatedAt := now.Add(-10 * time.Minute)

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          301,
		UserID:      1,
		ChallengeID: 201,
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   staleUpdatedAt,
		UpdatedAt:   staleUpdatedAt,
	})
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          302,
		UserID:      1,
		ChallengeID: 202,
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   staleUpdatedAt,
		UpdatedAt:   staleUpdatedAt,
	})
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          303,
		UserID:      1,
		ChallengeID: 203,
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now.Add(-time.Minute),
		UpdatedAt:   now.Add(-time.Minute),
	})
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          304,
		UserID:      1,
		ChallengeID: 204,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   staleUpdatedAt,
		UpdatedAt:   staleUpdatedAt,
	})

	instances, err := repo.ListStoppingInstances(context.Background(), cutoff)
	if err != nil {
		t.Fatalf("ListStoppingInstances() error = %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 stale stopping instances, got %d", len(instances))
	}
	if instances[0].ID != 301 || instances[1].ID != 302 {
		t.Fatalf("expected stale stopping instances ordered by updated_at then id, got ids=%d,%d", instances[0].ID, instances[1].ID)
	}
}

func TestServiceCreateContainerCreatesIsolatedNetwork(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:           "net-123",
		containerID:         "ctr-123",
		resolvedServicePort: 80,
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	containerID, networkID, hostPort, servicePort, err := service.CreateContainer(context.Background(), "ctf/web:v1", map[string]string{"FLAG": "flag{1}"}, 0)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	if containerID != "ctr-123" {
		t.Fatalf("unexpected container id: %s", containerID)
	}
	if networkID != "net-123" {
		t.Fatalf("unexpected network id: %s", networkID)
	}
	if hostPort != 30000 {
		t.Fatalf("unexpected host port: %d", hostPort)
	}
	if servicePort != 80 {
		t.Fatalf("unexpected service port: %d", servicePort)
	}
	if engine.createdNetworkName == "" {
		t.Fatalf("expected isolated network to be created")
	}
	if engine.createdContainerCfg == nil || engine.createdContainerCfg.Network != engine.createdNetworkName {
		t.Fatalf("expected container to join created network, cfg=%+v network=%s", engine.createdContainerCfg, engine.createdNetworkName)
	}
	if _, exists := engine.createdContainerCfg.Ports["80"]; !exists {
		t.Fatalf("expected container to publish resolved service port 80, got %+v", engine.createdContainerCfg.Ports)
	}
	if got := engine.createdContainerCfg.Labels[runtimedomain.ComposeProjectLabelKey]; got != runtimedomain.ProjectLabelValue {
		t.Fatalf("expected compose project label %q, got %q", runtimedomain.ProjectLabelValue, got)
	}
	if got := engine.createdContainerCfg.Labels[runtimedomain.ComposeServiceLabelKey]; got != runtimedomain.ComposeServiceJeopardy {
		t.Fatalf("expected jeopardy compose service label, got %q", got)
	}
	if got := engine.createdNetworkLabel[runtimedomain.ComposeServiceLabelKey]; got != runtimedomain.ComposeServiceJeopardy {
		t.Fatalf("expected jeopardy network label, got %q", got)
	}
	if engine.createdNetworkSubnet != "10.11.0.0/29" {
		t.Fatalf("expected first single-container subnet 10.11.0.0/29, got %q", engine.createdNetworkSubnet)
	}
}

func TestServiceCreateContainerReservesAllocatedHostPort(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:           "net-reserve",
		containerID:         "ctr-reserve",
		resolvedServicePort: 80,
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	_, _, hostPort, _, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil, 0)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", hostPort).Count(&count).Error; err != nil {
		t.Fatalf("count reserved port allocation: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected host port %d to be reserved once, count=%d", hostPort, count)
	}
}

func TestRuntimeCleanupServiceReleasesRuntimeDetailHostPort(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{}
	service := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)
	now := time.Now()
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:      30001,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{
				ContainerID:  "ctr-cleanup",
				HostPort:     30001,
				IsEntryPoint: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := service.CleanupRuntime(context.Background(), &instanceentity.Instance{RuntimeDetails: runtimeDetails}); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30001).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected runtime detail host port to be released, count=%d", count)
	}
}

func TestRuntimeCleanupServiceReleasesOwnedRuntimeDetailHostPort(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{}
	service := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)
	now := time.Now()
	instanceID := int64(3201)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30011,
		InstanceID: &instanceID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create owned port allocation: %v", err)
	}
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{
				ContainerID:  "ctr-cleanup-owned",
				HostPort:     30011,
				IsEntryPoint: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := service.CleanupRuntime(context.Background(), &instanceentity.Instance{ID: instanceID, RuntimeDetails: runtimeDetails}); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30011).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected owned runtime detail host port to be released, count=%d", count)
	}
}

func TestRuntimeCleanupServiceReleasesOwnedRuntimeDetailSubnet(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{}
	service := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)
	now := time.Now()
	instanceID := int64(3202)
	if err := repo.db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.5.0/24",
		InstanceID: &instanceID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create owned subnet allocation: %v", err)
	}
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Networks: []runtimecontracts.InstanceRuntimeNetwork{
			{
				Key:       runtimecontracts.TopologyDefaultNetworkKey,
				NetworkID: "net-owned-subnet",
				Subnet:    "10.10.5.0/24",
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := service.CleanupRuntime(context.Background(), &instanceentity.Instance{ID: instanceID, RuntimeDetails: runtimeDetails}); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.NetworkAllocation{}).Where("subnet = ?", "10.10.5.0/24").Count(&count).Error; err != nil {
		t.Fatalf("count subnet allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected owned runtime detail subnet to be released, count=%d", count)
	}
}

func TestRuntimeCleanupServiceKeepsForeignOwnedRuntimeDetailHostPort(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{}
	service := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)
	now := time.Now()
	ownerInstanceID := int64(3202)
	otherInstanceID := int64(3203)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30012,
		InstanceID: &otherInstanceID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create foreign-owned port allocation: %v", err)
	}
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{
				ContainerID:  "ctr-cleanup-foreign",
				HostPort:     30012,
				IsEntryPoint: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := service.CleanupRuntime(context.Background(), &instanceentity.Instance{ID: ownerInstanceID, RuntimeDetails: runtimeDetails}); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30012).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected foreign-owned runtime detail host port to stay allocated, count=%d", count)
	}
}

func TestServiceCreateContainerFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := runtimecmd.NewProvisioningService(repo, nil, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
		PublicHost:         "127.0.0.1",
	}, nil)

	containerID, networkID, hostPort, servicePort, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil, 0)
	if err == nil {
		t.Fatal("expected CreateContainer() to fail when runtime engine is unavailable")
	}
	if !errors.Is(err, runtimeports.ErrRuntimeEngineUnavailable) {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
	if containerID != "" || networkID != "" || hostPort != 0 || servicePort != 0 {
		t.Fatalf("expected zero runtime result on failure, got container=%q network=%q hostPort=%d servicePort=%d", containerID, networkID, hostPort, servicePort)
	}
}

func TestServiceCreateContainerRemovesNetworkWhenStartFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:   "net-456",
		containerID: "ctr-456",
		startErr:    errors.New("start failed"),
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	_, _, _, _, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil, 0)
	if err == nil {
		t.Fatal("expected CreateContainer() to fail")
	}
	if engine.removedContainerID != "ctr-456" {
		t.Fatalf("expected container cleanup, got %s", engine.removedContainerID)
	}
	if engine.removedNetworkID != "net-456" {
		t.Fatalf("expected network cleanup, got %s", engine.removedNetworkID)
	}
	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30000).Count(&count).Error; err != nil {
		t.Fatalf("count released reserved port allocation: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected reserved port rollback cleanup, count=%d", count)
	}
}

func TestServiceRemoveContainerFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	cleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil, nil)

	err := cleanupService.RemoveContainer(context.Background(), "ctr-missing-engine")
	if err == nil {
		t.Fatal("expected RemoveContainer() to fail when runtime engine is unavailable")
	}
	if !errors.Is(err, runtimeports.ErrRuntimeEngineUnavailable) {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
}

func TestServiceRemoveContainerHonorsCancellation(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			if containerID != "ctr-ctx" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cleanupService.RemoveContainer(ctx, "ctr-ctx"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceRemoveContainerIgnoresMissingContainer(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerErr: runtimeports.WrapRuntimeContainerNotFound(errors.New("Error response from daemon: No such container: ctr-missing")),
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	if err := cleanupService.RemoveContainer(context.Background(), "ctr-missing"); err != nil {
		t.Fatalf("expected missing container to be ignored, got %v", err)
	}
	if engine.removedContainerID != "ctr-missing" {
		t.Fatalf("expected cleanup to attempt removing ctr-missing, got %s", engine.removedContainerID)
	}
}

func TestServiceCleanupRuntimeFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	cleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil, nil)
	instance := &instanceentity.Instance{
		ID:          3002,
		ContainerID: "ctr-missing-engine",
		NetworkID:   "net-missing-engine",
	}

	err := cleanupService.CleanupRuntime(context.Background(), instance)
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

	if err := cleanupService.CleanupRuntime(ctx, instance); !errors.Is(err, context.Canceled) {
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

	if err := cleanupService.CleanupRuntime(context.Background(), instance); err != nil {
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

	if err := cleanupService.CleanupRuntime(context.Background(), instance); err != nil {
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

	if err := cleanupService.CleanupRuntime(context.Background(), instance); err != nil {
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

	if err := cleanupService.CleanupRuntime(context.Background(), instance); err != nil {
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

func TestServiceDestroyInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(301)
	teamID := int64(401)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Alpha", CaptainID: 1, InviteCode: "alpha", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          901,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 101,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	if err := service.DestroyInstance(context.Background(), 901, 2); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}

	instance, err := repo.FindByID(context.Background(), 901)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != instanceentity.InstanceStatusStopping {
		t.Fatalf("expected stopping status, got %s", instance.Status)
	}
}

func TestServiceExtendInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(302)
	teamID := int64(402)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Beta", CaptainID: 1, InviteCode: "beta", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	initialExpiry := now.Add(time.Hour)
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          902,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		ContainerID: "contest-shared-extend",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   initialExpiry,
	})

	resp, err := service.ExtendInstance(context.Background(), 902, 2)
	if err != nil {
		t.Fatalf("ExtendInstance() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected extend response")
	}
	if resp.RemainingExtends != 1 {
		t.Fatalf("expected remaining extends 1, got %d", resp.RemainingExtends)
	}

	instance, err := repo.FindByID(context.Background(), 902)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if !instance.ExpiresAt.After(initialExpiry) {
		t.Fatalf("expected expiry to be extended, got %s", instance.ExpiresAt)
	}
	if instance.ExtendCount != 1 {
		t.Fatalf("expected extend count 1, got %d", instance.ExtendCount)
	}
}

func TestServiceDestroyInstanceRejectsAWDTeamServiceInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(303)
	teamID := int64(403)
	serviceID := int64(503)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Gamma", CaptainID: 1, InviteCode: "gamma", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          905,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 105,
		ServiceID:   &serviceID,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	err := service.DestroyInstance(context.Background(), 905, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for awd team service destroy, got %v", err)
	}
}

func TestServiceExtendInstanceRejectsAWDTeamServiceInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(304)
	teamID := int64(404)
	serviceID := int64(504)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Delta", CaptainID: 1, InviteCode: "delta", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          906,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 106,
		ServiceID:   &serviceID,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	_, err := service.ExtendInstance(context.Background(), 906, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for awd team service extend, got %v", err)
	}
}

func TestServiceDestroyInstanceRejectsSharedInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:              903,
		Title:           "Shared Practice",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      taxonomy.DifficultyEasy,
		FlagType:        challengecontracts.FlagTypeStatic,
		Status:          challengecontracts.ChallengeStatusPublished,
		InstanceSharing: challengecontracts.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          903,
		UserID:      1,
		ChallengeID: 903,
		ShareScope:  instancecontracts.ShareScopeShared,
		ContainerID: "shared-ctr",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	err := service.DestroyInstance(context.Background(), 903, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for shared destroy, got %v", err)
	}
}

func TestServiceExtendInstanceRejectsSharedInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:              904,
		Title:           "Shared Practice",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      taxonomy.DifficultyEasy,
		FlagType:        challengecontracts.FlagTypeStatic,
		Status:          challengecontracts.ChallengeStatusPublished,
		InstanceSharing: challengecontracts.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          904,
		UserID:      1,
		ChallengeID: 904,
		ShareScope:  instancecontracts.ShareScopeShared,
		ContainerID: "shared-ctr",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	_, err := service.ExtendInstance(context.Background(), 904, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for shared extend, got %v", err)
	}
}

func TestServiceGetUserInstancesIncludesChallengeMetadata(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:         101,
		Title:      "Matrix Web Challenge",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		FlagType:   challengecontracts.FlagTypeStatic,
		Status:     challengecontracts.ChallengeStatusPublished,
		Points:     100,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1001,
		UserID:      1,
		ChallengeID: 101,
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30001",
		ExpiresAt:   now.Add(time.Hour),
		ExtendCount: 1,
		MaxExtends:  3,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 instance, got %+v", items)
	}
	item := items[0]
	if item.ChallengeTitle != "Matrix Web Challenge" {
		t.Fatalf("expected challenge title, got %+v", item)
	}
	if item.Category != taxonomy.DimensionWeb {
		t.Fatalf("expected category %q, got %+v", taxonomy.DimensionWeb, item)
	}
	if item.Difficulty != taxonomy.DifficultyEasy {
		t.Fatalf("expected difficulty %q, got %+v", taxonomy.DifficultyEasy, item)
	}
	if item.FlagType != challengecontracts.FlagTypeStatic {
		t.Fatalf("expected flag type %q, got %+v", challengecontracts.FlagTypeStatic, item)
	}
	if item.RemainingExtends != 2 {
		t.Fatalf("expected remaining extends 2, got %+v", item)
	}
}

func TestServiceGetUserInstancesShowsContestSharedInstanceToTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(501)
	teamID := int64(601)

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:         102,
		Title:      "Shared AWD Challenge",
		Category:   taxonomy.DimensionPwn,
		Difficulty: taxonomy.DifficultyMedium,
		FlagType:   challengecontracts.FlagTypeDynamic,
		Status:     challengecontracts.ChallengeStatusPublished,
		Points:     150,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime Team",
		CaptainID:  1,
		InviteCode: "runtime",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1002,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30002",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1002 {
		t.Fatalf("expected teammate visible shared instance, got %+v", items)
	}
}

func TestServiceGetUserInstancesShowsPracticeSharedInstanceToAnyUser(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:              103,
		Title:           "Shared Practice",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      taxonomy.DifficultyEasy,
		FlagType:        challengecontracts.FlagTypeStatic,
		Status:          challengecontracts.ChallengeStatusPublished,
		InstanceSharing: challengecontracts.InstanceSharingShared,
		Points:          80,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1003,
		UserID:      1,
		ChallengeID: 103,
		ShareScope:  instancecontracts.ShareScopeShared,
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30003",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1003 {
		t.Fatalf("expected global shared instance visible to another user, got %+v", items)
	}
	if items[0].ShareScope != instancecontracts.ShareScopeShared {
		t.Fatalf("expected share scope to be returned, got %+v", items[0])
	}
}

func TestServiceCreateTopologyCreatesMultipleContainersOnSharedNetwork(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-789",
		containerIDs: []string{"web-ctr", "db-ctr"},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.PrimaryContainerID != "web-ctr" || result.NetworkID != "net-789" {
		t.Fatalf("unexpected topology result: %+v", result)
	}
	if len(result.RuntimeDetails.Containers) != 2 {
		t.Fatalf("unexpected runtime details: %+v", result.RuntimeDetails)
	}
	if len(engine.createdContainerCfgs) != 2 {
		t.Fatalf("expected two create container calls, got %d", len(engine.createdContainerCfgs))
	}
	if engine.createdContainerCfgs[0].Network != engine.createdNetworkName || engine.createdContainerCfgs[1].Network != engine.createdNetworkName {
		t.Fatalf("expected all containers to join shared network")
	}
	if engine.createdNetworkSubnet != "10.10.0.0/24" {
		t.Fatalf("expected topology subnet 10.10.0.0/24, got %q", engine.createdNetworkSubnet)
	}
	if engine.createdNetworkAllowExisting {
		t.Fatal("non-shared topology network must not reuse an existing Docker network")
	}
	if _, exists := engine.createdContainerCfgs[1].Ports["8080"]; exists {
		t.Fatalf("non-entry container should not expose host port")
	}
}

func TestServiceCreateTopologyCanKeepEntryPointPrivate(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-private",
		containerIDs: []string{"web-private"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-private" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{engine.createdNetworkName: "172.30.0.10"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.AccessURL != "http://172.30.0.10:8080" {
		t.Fatalf("expected private access url, got %q", result.AccessURL)
	}
	if len(engine.createdContainerCfgs) != 1 {
		t.Fatalf("expected one create container call, got %d", len(engine.createdContainerCfgs))
	}
	if len(engine.createdContainerCfgs[0].Ports) != 0 {
		t.Fatalf("entry container should not publish host port, got %+v", engine.createdContainerCfgs[0].Ports)
	}
	if got := result.RuntimeDetails.Containers[0].HostPort; got != 0 {
		t.Fatalf("expected no runtime host port, got %d", got)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no reserved host ports, count=%d", count)
	}
}

func TestServiceCreateTopologyUsesPreferredContainerName(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-named",
		containerIDs: []string{"web-named"},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	preferredName := "ctf-instance-bank-portal-c8-t15"
	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		ContainerName: preferredName,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if engine.createdContainerCfg == nil {
		t.Fatal("expected container config to be created")
	}
	if engine.createdContainerCfg.Name != preferredName {
		t.Fatalf("expected preferred container name %q, got %q", preferredName, engine.createdContainerCfg.Name)
	}
}

func TestServiceCreateContainerMarksAWDImagesAsAWDComposeService(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:           "net-awd-image",
		containerID:         "ctr-awd-image",
		resolvedServicePort: 8080,
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	if _, _, _, _, err := service.CreateContainer(context.Background(), "127.0.0.1:5000/awd/awd-supply-ticket:latest", nil, 0); err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	if got := engine.createdContainerCfg.Labels[runtimedomain.ComposeServiceLabelKey]; got != runtimedomain.ComposeServiceAWD {
		t.Fatalf("expected awd compose service label, got %q", got)
	}
	if got := engine.createdNetworkLabel[runtimedomain.ComposeServiceLabelKey]; got != runtimedomain.ComposeServiceAWD {
		t.Fatalf("expected awd network label, got %q", got)
	}
}

func TestServiceCreateTopologyMarksAWDWorkspaceAsAWDComposeService(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-awd-workspace",
		containerIDs: []string{"workspace-ctr"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			return map[string]string{engine.createdNetworkName: "172.30.0.44"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		ContainerName:              "ctf-workspace-workspace-c8-t15-s21-r2",
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:             "workspace",
				Image:           "python:3.12-alpine",
				ServicePort:     22,
				ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
				NetworkAliases:  []string{"awd-ws-c8-t15-s21-r2"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if got := engine.createdContainerCfg.Labels[runtimedomain.ComposeServiceLabelKey]; got != runtimedomain.ComposeServiceAWD {
		t.Fatalf("expected awd compose service label, got %q", got)
	}
	if got := engine.createdNetworkLabel[runtimedomain.ComposeServiceLabelKey]; got != runtimedomain.ComposeServiceAWD {
		t.Fatalf("expected awd network label, got %q", got)
	}
	if engine.createdNetworkSubnet != "" {
		t.Fatalf("expected shared AWD network to skip explicit subnet allocation, got %q", engine.createdNetworkSubnet)
	}
	if engine.listNetworkSubnetsCalls != 0 {
		t.Fatalf("expected shared-only topology to skip runtime subnet listing, got %d", engine.listNetworkSubnetsCalls)
	}
}

func TestServiceCreateTopologyPassesMountsAndCommandToEngine(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-mounts",
		containerIDs: []string{"workspace-ctr"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			return map[string]string{engine.createdNetworkName: "172.30.0.44"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:             "workspace",
				Image:           "python:3.12-alpine",
				Env:             map[string]string{"LANG": "C.UTF-8", "LC_ALL": "C.UTF-8", "TERM": "xterm-256color"},
				ServicePort:     22,
				ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
				WorkingDir:      "/workspace",
				Command:         []string{"tail", "-f", "/dev/null"},
				Mounts: []runtimecontracts.ContainerMount{
					{Source: "ctf-ws-src", Target: "/workspace/src"},
					{Source: "ctf-ws-data", Target: "/workspace/data", ReadOnly: true},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if engine.createdContainerCfg == nil {
		t.Fatal("expected container config to be created")
	}
	if engine.createdContainerCfg.WorkingDir != "/workspace" {
		t.Fatalf("expected working dir /workspace, got %q", engine.createdContainerCfg.WorkingDir)
	}
	if len(engine.createdContainerCfg.Command) != 3 || engine.createdContainerCfg.Command[0] != "tail" {
		t.Fatalf("expected workspace keepalive command, got %+v", engine.createdContainerCfg.Command)
	}
	if len(engine.createdContainerCfg.Mounts) != 2 {
		t.Fatalf("expected two mounts, got %+v", engine.createdContainerCfg.Mounts)
	}
	if engine.createdContainerCfg.Mounts[0].Source != "ctf-ws-src" || engine.createdContainerCfg.Mounts[0].Target != "/workspace/src" {
		t.Fatalf("unexpected writable mount: %+v", engine.createdContainerCfg.Mounts[0])
	}
	if engine.createdContainerCfg.Mounts[1].Source != "ctf-ws-data" || engine.createdContainerCfg.Mounts[1].Target != "/workspace/data" || !engine.createdContainerCfg.Mounts[1].ReadOnly {
		t.Fatalf("unexpected readonly mount: %+v", engine.createdContainerCfg.Mounts[1])
	}
	for _, wantEnv := range []string{"LANG=C.UTF-8", "LC_ALL=C.UTF-8", "TERM=xterm-256color"} {
		found := false
		for _, gotEnv := range engine.createdContainerCfg.Env {
			if gotEnv == wantEnv {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected env %q in container config, got %+v", wantEnv, engine.createdContainerCfg.Env)
		}
	}
}

func TestServiceCreateTopologyBuildsTCPEntryAccessURL(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-tcp",
		containerIDs: []string{"pwn-tcp"},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:             "pwn",
				Image:           "ctf/pwn:v1",
				ServicePort:     31337,
				ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.AccessURL != "tcp://127.0.0.1:30000" {
		t.Fatalf("expected tcp access url, got %q", result.AccessURL)
	}
	if got := result.RuntimeDetails.Containers[0].ServiceProtocol; got != challengecontracts.ChallengeTargetProtocolTCP {
		t.Fatalf("expected runtime details service protocol tcp, got %q", got)
	}
}

func TestServiceDestroyManagedInstanceMarksStoppingThenBackgroundCleanupRemovesRuntime(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{}
	service := newTestRuntimeModule(repo, engine)

	instance := &instanceentity.Instance{
		ID:             1,
		UserID:         1,
		ChallengeID:    1,
		HostPort:       30001,
		ContainerID:    "web-ctr",
		NetworkID:      "net-1",
		RuntimeDetails: `{"containers":[{"container_id":"web-ctr"},{"container_id":"db-ctr"}],"acl_rules":[{"comment":"ctf:acl:test","source_ip":"172.30.0.2","target_ip":"172.30.0.3","action":"allow","protocol":"tcp","ports":[3306]}]}`,
		Status:         instanceentity.InstanceStatusRunning,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&runtimeentity.PortAllocation{Port: 30001, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	if err := service.DestroyInstance(context.Background(), instance.ID, instance.UserID); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}
	if len(engine.removedContainerIDs) != 0 || len(engine.removedNetworkIDs) != 0 || len(engine.removedACLRules) != 0 {
		t.Fatalf("expected destroy request to return before runtime cleanup, got containers=%v networks=%v acl=%v", engine.removedContainerIDs, engine.removedNetworkIDs, engine.removedACLRules)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() after destroy error = %v", err)
	}
	if updated.Status != instanceentity.InstanceStatusStopping {
		t.Fatalf("expected instance to enter stopping before background cleanup, got %+v", updated)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunStoppingCleanupLoop(runCtx)
	}()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		updated, err = repo.FindByID(context.Background(), instance.ID)
		if err != nil {
			t.Fatalf("FindByID() during cleanup error = %v", err)
		}
		if updated.Status == instanceentity.InstanceStatusStopped {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done

	if len(engine.removedContainerIDs) != 2 {
		t.Fatalf("expected 2 removed containers, got %v", engine.removedContainerIDs)
	}
	if len(engine.removedNetworkIDs) != 1 || engine.removedNetworkIDs[0] != "net-1" {
		t.Fatalf("expected 1 removed network, got %v", engine.removedNetworkIDs)
	}
	if len(engine.removedACLRules) != 1 || engine.removedACLRules[0].Comment != "ctf:acl:test" {
		t.Fatalf("expected acl rules to be removed, got %+v", engine.removedACLRules)
	}

	if updated.Status != instanceentity.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %+v", updated)
	}
	if updated.HostPort != 0 || updated.ContainerID != "" || updated.NetworkID != "" || updated.RuntimeDetails != "" || updated.AccessURL != "" {
		t.Fatalf("expected stopped instance runtime fields to be cleared, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30001).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
}

func TestServiceCleanExpiredInstancesKeepsRunningStateWhenRuntimeCleanupFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instance := &instanceentity.Instance{
		ID:          2101,
		UserID:      1,
		ChallengeID: 1,
		HostPort:    30002,
		ContainerID: "web-ctr",
		NetworkID:   "net-2",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(-time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30002,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	engine := &fakeRuntimeEngine{removeContainerErr: errors.New("remove failed")}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, cleanupService, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	if err := service.CleanExpiredInstances(context.Background()); err != nil {
		t.Fatalf("CleanExpiredInstances() error = %v", err)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != instanceentity.InstanceStatusRunning {
		t.Fatalf("expected instance to remain running for retry, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30002).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected port allocation to remain reserved, count=%d", count)
	}
}

func TestServiceCleanExpiredInstancesMarksExpiredWhenContainerAlreadyRemoved(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instance := &instanceentity.Instance{
		ID:          2102,
		UserID:      1,
		ChallengeID: 1,
		HostPort:    30003,
		ContainerID: "missing-ctr",
		NetworkID:   "net-3",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(-time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30003,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	engine := &fakeRuntimeEngine{
		removeContainerErr: runtimeports.WrapRuntimeContainerNotFound(errors.New("Error response from daemon: No such container: missing-ctr")),
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)
	service := instancecmd.NewInstanceMaintenanceService(repo, nil, cleanupService, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	if err := service.CleanExpiredInstances(context.Background()); err != nil {
		t.Fatalf("CleanExpiredInstances() error = %v", err)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != instanceentity.InstanceStatusExpired {
		t.Fatalf("expected instance to be marked expired, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30003).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
}

func TestRepositoryRequeueLostRuntimePreservesInstanceScope(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	contestID := int64(3101)
	teamID := int64(4101)
	serviceID := int64(7101)
	instance := &instanceentity.Instance{
		ID:             2201,
		UserID:         5101,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    6101,
		ServiceID:      &serviceID,
		HostPort:       30004,
		ContainerID:    "lost-container",
		NetworkID:      "lost-network",
		RuntimeDetails: `{"containers":[{"container_id":"lost-container"}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://10.10.0.2:8080",
		Nonce:          "nonce-2201",
		ExpiresAt:      now.Add(time.Hour),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&runtimeentity.PortAllocation{
		Port:       30004,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	requeued, err := repo.RequeueLostRuntime(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("RequeueLostRuntime() error = %v", err)
	}
	if !requeued {
		t.Fatal("expected instance to be requeued")
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status, got %+v", updated)
	}
	if updated.ContainerID != "" || updated.NetworkID != "" || updated.RuntimeDetails != "" || updated.AccessURL != "" {
		t.Fatalf("expected runtime fields cleared, got %+v", updated)
	}
	if updated.UserID != instance.UserID || updated.ChallengeID != instance.ChallengeID || updated.ShareScope != instanceentity.ShareScopePerTeam || updated.Nonce != instance.Nonce || updated.HostPort != instance.HostPort {
		t.Fatalf("expected instance scope preserved, got %+v", updated)
	}
	if updated.ContestID == nil || *updated.ContestID != contestID || updated.TeamID == nil || *updated.TeamID != teamID || updated.ServiceID == nil || *updated.ServiceID != serviceID {
		t.Fatalf("expected contest/team/service scope preserved, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&runtimeentity.PortAllocation{}).Where("port = ?", 30004).Count(&count).Error; err != nil {
		t.Fatalf("count port allocation: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected port allocation to remain reserved, count=%d", count)
	}
}

func TestServiceCreateTopologyUsesStableAliasForPrivateEntryPoint(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-awd-contest-8",
		containerIDs: []string{"web-awd"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-awd" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{"ctf-awd-contest-8": "172.30.0.20"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:            "web",
				Image:          "ctf/web:v1",
				ServicePort:    8080,
				IsEntryPoint:   true,
				NetworkKeys:    []string{runtimecontracts.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c8-t15-s21"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.AccessURL != "http://awd-c8-t15-s21:8080" {
		t.Fatalf("expected alias access url, got %q", result.AccessURL)
	}
	if len(result.RuntimeDetails.Networks) != 1 || !result.RuntimeDetails.Networks[0].Shared || result.RuntimeDetails.Networks[0].Name != "ctf-awd-contest-8" {
		t.Fatalf("expected shared contest network details, got %+v", result.RuntimeDetails.Networks)
	}
	if len(result.RuntimeDetails.Containers) != 1 || len(result.RuntimeDetails.Containers[0].NetworkAliases) != 1 || result.RuntimeDetails.Containers[0].NetworkAliases[0] != "awd-c8-t15-s21" {
		t.Fatalf("expected runtime alias details, got %+v", result.RuntimeDetails.Containers)
	}
	if result.RuntimeDetails.Containers[0].NetworkIPs["ctf-awd-contest-8"] != "172.30.0.20" {
		t.Fatalf("expected runtime network ip details, got %+v", result.RuntimeDetails.Containers[0].NetworkIPs)
	}
	if engine.createdContainerCfg == nil || len(engine.createdContainerCfg.NetworkAliases) != 1 || engine.createdContainerCfg.NetworkAliases[0] != "awd-c8-t15-s21" {
		t.Fatalf("expected Docker network alias in container config, got %+v", engine.createdContainerCfg)
	}
	if !engine.createdNetworkAllowExisting {
		t.Fatal("shared AWD contest network should allow reusing an existing Docker network")
	}
}

func TestServiceCreateTopologyCreatesAndConnectsMultipleNetworks(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkIDs:   []string{"net-public", "net-backend"},
		containerIDs: []string{"web-ctr", "db-ctr"},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: "public"},
			{Key: "backend", Internal: true},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{"public", "backend"}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{"backend"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.NetworkID != "net-public" {
		t.Fatalf("unexpected primary network id: %+v", result)
	}
	if len(result.RuntimeDetails.Networks) != 2 {
		t.Fatalf("unexpected runtime networks: %+v", result.RuntimeDetails)
	}
	if len(engine.createdNetworkNames) != 2 {
		t.Fatalf("expected two created networks, got %v", engine.createdNetworkNames)
	}
	if len(engine.connectedNetworks["web-ctr"]) != 1 || engine.connectedNetworks["web-ctr"][0] != engine.createdNetworkNames[1] {
		t.Fatalf("expected web container to connect to backend network, got %+v", engine.connectedNetworks)
	}
	if len(engine.connectedNetworks["db-ctr"]) != 0 {
		t.Fatalf("db container should not need extra network connect, got %+v", engine.connectedNetworks)
	}
	if len(engine.createdNetworkSubnets) != 2 {
		t.Fatalf("expected two explicit network subnets, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[0] == "" || engine.createdNetworkSubnets[1] == "" {
		t.Fatalf("expected explicit subnets for non-shared topology networks, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[0] == engine.createdNetworkSubnets[1] {
		t.Fatalf("expected distinct subnets per runtime network, got %+v", engine.createdNetworkSubnets)
	}
}

func TestServiceCreateTopologyLogsProvisioningStages(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkIDs:          []string{"net-stage-primary", "net-stage-extra"},
		containerIDs:        []string{"web-stage", "worker-stage"},
		resolvedServicePort: 8080,
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			switch containerID {
			case "web-stage":
				return map[string]string{
					engine.createdNetworkNames[0]: "172.32.0.10",
					engine.createdNetworkNames[1]: "172.32.1.10",
				}
			case "worker-stage":
				return map[string]string{
					engine.createdNetworkNames[1]: "172.32.1.20",
				}
			default:
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return nil
		},
	}
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, logger)

	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: 4242,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
			{Key: "extra"},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:          "web",
				Image:        "ctf/web:v1",
				IsEntryPoint: true,
				NetworkKeys:  []string{runtimecontracts.TopologyDefaultNetworkKey, "extra"},
			},
			{
				Key:         "worker",
				Image:       "ctf/worker:v1",
				NetworkKeys: []string{"extra"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}

	entries := observed.FilterMessage("runtime provisioning stage succeeded").AllUntimed()
	stageCounts := make(map[string]int, len(entries))
	var (
		entryCreateLog  map[string]any
		workerCreateLog map[string]any
	)
	for _, entry := range entries {
		ctxMap := entry.ContextMap()
		stage, _ := ctxMap["stage"].(string)
		stageCounts[stage]++
		if got, ok := ctxMap["instance_id"].(int64); !ok || got != 4242 {
			t.Fatalf("expected instance_id=4242 in stage log, got %+v", ctxMap)
		}
		if stage == "container_create" {
			switch ctxMap["node_key"] {
			case "web":
				entryCreateLog = ctxMap
			case "worker":
				workerCreateLog = ctxMap
			}
		}
	}
	for _, stage := range []string{
		"network_create",
		"service_port_resolve",
		"container_create",
		"container_start",
		"connect_extra_networks",
		"inspect_network_ips",
	} {
		if stageCounts[stage] == 0 {
			t.Fatalf("expected stage %q to be logged, got counts=%v", stage, stageCounts)
		}
	}
	if entryCreateLog == nil || workerCreateLog == nil {
		t.Fatalf("expected container_create logs for entry and worker nodes, got entry=%v worker=%v", entryCreateLog, workerCreateLog)
	}
	if got, ok := entryCreateLog["host_port"].(int64); !ok || got <= 0 {
		t.Fatalf("expected entry container_create log to include host_port, got %+v", entryCreateLog)
	}
	if got, _ := entryCreateLog["container_id"].(string); got != "web-stage" {
		t.Fatalf("expected entry container_create log to include container_id=web-stage, got %+v", entryCreateLog)
	}
	if _, exists := workerCreateLog["host_port"]; exists {
		t.Fatalf("expected worker container_create log to omit host_port, got %+v", workerCreateLog)
	}
}

func TestServiceCreateTopologySkipsConflictingSubnetAndRetries(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:         "net-retried",
		containerIDs:      []string{"web-ctr"},
		createNetworkErrs: []error{runtimeports.WrapRuntimeNetworkSubnetConflict(errors.New("invalid pool request: Pool overlaps with other one on this address space"))},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-ctr" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{engine.createdNetworkNames[len(engine.createdNetworkNames)-1]: "172.30.0.20"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: 7001,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected topology result")
	}
	if len(engine.createdNetworkSubnets) != 2 {
		t.Fatalf("expected network create to retry with a new subnet, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[0] != "10.10.0.0/24" {
		t.Fatalf("expected first subnet attempt 10.10.0.0/24, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[1] != "10.10.1.0/24" {
		t.Fatalf("expected retry to skip conflicting subnet and use 10.10.1.0/24, got %+v", engine.createdNetworkSubnets)
	}
}

func TestServiceCreateTopologySkipsRuntimeOccupiedSubnetsBeforeCreate(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:          "net-fresh",
		containerIDs:       []string{"web-ctr"},
		listNetworkSubnets: []string{"10.10.0.0/24", "10.10.1.0/24"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-ctr" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{engine.createdNetworkNames[len(engine.createdNetworkNames)-1]: "172.30.0.20"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: 7002,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected topology result")
	}
	if len(engine.createdNetworkSubnets) != 1 {
		t.Fatalf("expected a single network create attempt after pre-filtering occupied subnets, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[0] != "10.10.2.0/24" {
		t.Fatalf("expected first free subnet after occupied runtime subnets to be 10.10.2.0/24, got %+v", engine.createdNetworkSubnets)
	}
}

func TestServiceCreateTopologySharesOccupiedSubnetsAcrossNetworks(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkIDs:          []string{"net-explicit", "net-dynamic"},
		containerIDs:        []string{"web-ctr"},
		listNetworkSubnets:  []string{"10.10.0.0/24"},
		resolvedServicePort: 8080,
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-ctr" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{
				engine.createdNetworkNames[0]: "172.30.1.10",
				engine.createdNetworkNames[1]: "172.30.2.10",
			}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: 7003,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Subnet: "10.10.1.0/24"},
			{Key: "backend"},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:          "web",
				Image:        "ctf/web:v1",
				IsEntryPoint: true,
				NetworkKeys:  []string{runtimecontracts.TopologyDefaultNetworkKey, "backend"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected topology result")
	}
	if engine.listNetworkSubnetsCalls != 1 {
		t.Fatalf("expected ListNetworkSubnets to be called once per topology, got %d", engine.listNetworkSubnetsCalls)
	}
	if len(engine.createdNetworkSubnets) != 2 {
		t.Fatalf("expected two created network subnets, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[0] != "10.10.1.0/24" {
		t.Fatalf("expected explicit subnet to be used first, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[1] != "10.10.2.0/24" {
		t.Fatalf("expected dynamic subnet to skip both runtime-occupied and topology-occupied subnets, got %+v", engine.createdNetworkSubnets)
	}
}

func TestServiceCreateTopologySkipsRuntimeSubnetListingForExplicitSubnetsOnly(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-explicit-only",
		containerIDs: []string{"web-ctr"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-ctr" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{engine.createdNetworkNames[len(engine.createdNetworkNames)-1]: "172.30.3.10"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: 7005,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Subnet: "10.10.10.0/24"},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected topology result")
	}
	if engine.listNetworkSubnetsCalls != 0 {
		t.Fatalf("expected explicit-subnet-only topology to skip runtime subnet listing, got %d", engine.listNetworkSubnetsCalls)
	}
	if len(engine.createdNetworkSubnets) != 1 || engine.createdNetworkSubnets[0] != "10.10.10.0/24" {
		t.Fatalf("expected explicit subnet to be used as-is, got %+v", engine.createdNetworkSubnets)
	}
}

func TestServiceCreateTopologySkipsRuntimeOccupiedOwnerReservationWithoutRetry(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	instanceID := int64(7004)
	now := time.Now()
	if err := repo.db.Create(&runtimeentity.NetworkAllocation{
		Subnet:     "10.10.9.0/24",
		InstanceID: &instanceID,
		NetworkKey: runtimecontracts.TopologyDefaultNetworkKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("seed owned subnet allocation: %v", err)
	}

	engine := &fakeRuntimeEngine{
		networkID:          "net-owner-refresh",
		containerIDs:       []string{"web-ctr"},
		listNetworkSubnets: []string{"10.10.9.0/24"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if containerID != "web-ctr" {
				t.Fatalf("unexpected inspect container id: %s", containerID)
			}
			return map[string]string{engine.createdNetworkNames[len(engine.createdNetworkNames)-1]: "172.30.0.21"}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: instanceID,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected topology result")
	}
	if len(engine.createdNetworkSubnets) != 1 {
		t.Fatalf("expected owner reservation refresh to avoid retry, got %+v", engine.createdNetworkSubnets)
	}
	if engine.createdNetworkSubnets[0] != "10.10.0.0/24" {
		t.Fatalf("expected runtime-occupied owner subnet to be reassigned before create, got %+v", engine.createdNetworkSubnets)
	}

	var allocation runtimeentity.NetworkAllocation
	if err := repo.db.Where("instance_id = ? AND network_key = ?", instanceID, runtimecontracts.TopologyDefaultNetworkKey).First(&allocation).Error; err != nil {
		t.Fatalf("load updated subnet allocation: %v", err)
	}
	if allocation.Subnet != "10.10.0.0/24" {
		t.Fatalf("expected owner allocation to update to 10.10.0.0/24, got %q", allocation.Subnet)
	}
}

func TestServiceCreateTopologyLogsStageFailure(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		createNetworkErrs: []error{context.DeadlineExceeded},
	}
	core, observed := observer.New(zap.WarnLevel)
	logger := zap.New(core)
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, logger)

	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		OwnerInstanceID: 5252,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	entries := observed.FilterMessage("runtime provisioning stage failed").AllUntimed()
	if len(entries) != 1 {
		t.Fatalf("expected one failed stage log, got %d", len(entries))
	}
	ctxMap := entries[0].ContextMap()
	if got, _ := ctxMap["stage"].(string); got != "network_create" {
		t.Fatalf("expected network_create failure stage, got %+v", ctxMap)
	}
	if got, _ := ctxMap["network_key"].(string); got != runtimecontracts.TopologyDefaultNetworkKey {
		t.Fatalf("expected default network key in failure log, got %+v", ctxMap)
	}
	if got, ok := ctxMap["instance_id"].(int64); !ok || got != 5252 {
		t.Fatalf("expected instance_id=5252 in failure log, got %+v", ctxMap)
	}
	if _, exists := ctxMap["error"]; !exists {
		t.Fatalf("expected failure log to include error field, got %+v", ctxMap)
	}
}

func TestServiceCreateTopologyAppliesFineGrainedACLRules(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-acl",
		containerIDs: []string{"web-ctr", "db-ctr"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if len(engine.createdNetworkNames) == 0 {
				return nil
			}
			switch containerID {
			case "web-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.30.0.2"}
			case "db-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.30.0.3"}
			default:
				return nil
			}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
		Policies: []runtimecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: runtimecontracts.TopologyPolicyActionAllow, Protocol: runtimecontracts.TopologyPolicyProtocolTCP, Ports: []int{3306}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if len(engine.appliedACLRules) != 2 {
		t.Fatalf("expected 2 acl rules, got %+v", engine.appliedACLRules)
	}
	if len(result.RuntimeDetails.ACLRules) != 2 {
		t.Fatalf("expected runtime acl rules, got %+v", result.RuntimeDetails.ACLRules)
	}
	if engine.appliedACLRules[0].Action != runtimecontracts.TopologyPolicyActionAllow || engine.appliedACLRules[0].Protocol != runtimecontracts.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected allow acl rule: %+v", engine.appliedACLRules[0])
	}
	if engine.appliedACLRules[1].Action != runtimecontracts.TopologyPolicyActionDeny || len(engine.appliedACLRules[1].Ports) != 0 {
		t.Fatalf("unexpected fallback deny rule: %+v", engine.appliedACLRules[1])
	}
}

func TestServiceCreateTopologyRollsBackWhenACLApplyFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-rollback",
		containerIDs: []string{"web-ctr", "db-ctr"},
		applyACLErr:  errors.New("acl apply failed"),
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if len(engine.createdNetworkNames) == 0 {
				return nil
			}
			switch containerID {
			case "web-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.31.0.2"}
			case "db-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.31.0.3"}
			default:
				return nil
			}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
		Policies: []runtimecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: runtimecontracts.TopologyPolicyActionAllow, Protocol: runtimecontracts.TopologyPolicyProtocolTCP, Ports: []int{3306}},
		},
	})
	if err == nil {
		t.Fatal("expected CreateTopology() to fail")
	}
	if len(engine.removedContainerIDs) != 2 {
		t.Fatalf("expected created containers to be cleaned up, got %v", engine.removedContainerIDs)
	}
	if len(engine.removedNetworkIDs) != 1 || engine.removedNetworkIDs[0] != "net-rollback" {
		t.Fatalf("expected created network to be cleaned up, got %v", engine.removedNetworkIDs)
	}
}

func TestServiceListTeacherInstancesScopesTeacherAndAppliesFilters(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &identitycontracts.User{ID: 1, Username: "teacher-a", Role: identitycontracts.RoleTeacher, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: identitycontracts.RoleStudent, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 3, Username: "bob", StudentNo: "S-1002", Role: identitycontracts.RoleStudent, ClassName: "Class B", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &runtimeChallengeTestRow{ID: 11, Title: "web-101", Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 101, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 102, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 103, UserID: 2, ChallengeID: 11, ContainerID: "inst-stopped", Status: instanceentity.InstanceStatusStopped, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})

	pageResp, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{})
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(pageResp.List) != 1 {
		t.Fatalf("expected 1 visible instance, got %d (%+v)", len(pageResp.List), pageResp.List)
	}
	if pageResp.List[0].StudentUsername != "alice" || pageResp.List[0].ClassName != "Class A" {
		t.Fatalf("unexpected item: %+v", pageResp.List[0])
	}

	filtered, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{
		Keyword:   "ali",
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with filters error = %v", err)
	}
	if len(filtered.List) != 1 || filtered.List[0].ID != 101 {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}

	byStudentNoKeyword, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{
		Keyword: "1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with student_no keyword error = %v", err)
	}
	if len(byStudentNoKeyword.List) != 1 || byStudentNoKeyword.List[0].ID != 101 {
		t.Fatalf("expected keyword to match student_no, got %+v", byStudentNoKeyword)
	}
}

func TestServiceListTeacherInstancesRejectsTeacherCrossClassFilter(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &identitycontracts.User{ID: 1, Username: "teacher-a", Role: identitycontracts.RoleTeacher, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})

	_, err := service.ListTeacherInstances(context.Background(), 1, identitycontracts.RoleTeacher, instancecontracts.TeacherInstanceListQuery{ClassName: "Class B"})
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceDestroyTeacherInstanceHonorsClassScope(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &identitycontracts.User{ID: 1, Username: "teacher-a", Role: identitycontracts.RoleTeacher, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 2, Username: "alice", Role: identitycontracts.RoleStudent, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &identitycontracts.User{ID: 3, Username: "bob", Role: identitycontracts.RoleStudent, ClassName: "Class B", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &runtimeChallengeTestRow{ID: 11, Title: "web-101", Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 201, UserID: 2, ChallengeID: 11, Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &instanceentity.Instance{ID: 202, UserID: 3, ChallengeID: 11, Status: instanceentity.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, identitycontracts.RoleTeacher); err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, identitycontracts.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	instance, err := repo.FindByID(context.Background(), 201)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != instanceentity.InstanceStatusStopping {
		t.Fatalf("expected stopping status, got %s", instance.Status)
	}
}

type runtimeTestRepository struct {
	*runtimeinfra.Repository
	db *gorm.DB
}

func newTestRepository(t *testing.T) *runtimeTestRepository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&identitycontracts.User{}, &runtimeChallengeTestRow{}, &instanceentity.Instance{}, &runtimeentity.PortAllocation{}, &runtimeentity.NetworkAllocation{}, &contestcontracts.ContestRegistration{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&contestcontracts.Team{}, &contestcontracts.TeamMember{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&contestcontracts.Contest{}, &contestcontracts.ContestAWDService{}); err != nil {
		t.Fatalf("migrate awd tables: %v", err)
	}
	if err := db.AutoMigrate(&runtimecontracts.AWDScopeControl{}); err != nil {
		t.Fatalf("migrate awd scope control tables: %v", err)
	}
	if err := db.AutoMigrate(&runtimeentity.AWDServiceOperation{}); err != nil {
		t.Fatalf("migrate awd operation tables: %v", err)
	}
	return &runtimeTestRepository{
		Repository: runtimeinfra.NewRepository(db),
		db:         db,
	}
}

type testRuntimeService struct {
	commands    *instancecmd.InstanceService
	queries     *instanceqry.InstanceService
	maintenance *instancecmd.InstanceMaintenanceService
}

func (s *testRuntimeService) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	return s.commands.DestroyInstance(ctx, instanceID, userID)
}

func (s *testRuntimeService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*instancecontracts.InstanceResp, error) {
	return s.commands.ExtendInstance(ctx, instanceID, userID)
}

func (s *testRuntimeService) GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error) {
	return s.queries.GetUserInstances(ctx, userID)
}

func (s *testRuntimeService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	return s.queries.GetAccessURL(ctx, instanceID, userID)
}

func (s *testRuntimeService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query instancecontracts.TeacherInstanceListQuery) (*instancecontracts.TeacherInstancePageResult, error) {
	return s.queries.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (s *testRuntimeService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	return s.commands.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (s *testRuntimeService) RunStoppingCleanupLoop(ctx context.Context) {
	if s == nil || s.maintenance == nil {
		return
	}
	s.maintenance.RunStoppingCleanupLoop(ctx)
}

func newTestRuntimeModule(repo *runtimeTestRepository, engine *fakeRuntimeEngine) *testRuntimeService {
	cfg := &config.ContainerConfig{
		MaxExtends:          2,
		ExtendDuration:      30 * time.Minute,
		OrphanGracePeriod:   5 * time.Minute,
		DeletePollInterval:  5 * time.Millisecond,
		DeleteMaxConcurrent: 2,
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)
	return &testRuntimeService{
		commands:    instancecmd.NewInstanceService(repo, cleanupService, cfg, nil),
		queries:     instanceqry.NewInstanceService(repo, cfg),
		maintenance: instancecmd.NewInstanceMaintenanceService(repo, nil, cleanupService, cfg, nil),
	}
}

type fakeRuntimeEngine struct {
	networkID                      string
	networkIDs                     []string
	createNetworkErrs              []error
	listNetworkSubnets             []string
	listNetworkSubnetsErr          error
	listNetworkSubnetsCalls        int
	containerID                    string
	containerIDs                   []string
	startErr                       error
	applyACLErr                    error
	removeContainerErr             error
	removeNetworkErr               error
	resolvedServicePort            int
	resolveServicePortErr          error
	createdNetworkName             string
	createdNetworkNames            []string
	createdNetworkAllowExisting    bool
	createdNetworkAllowExistingSeq []bool
	createdNetworkLabel            map[string]string
	createdNetworkSubnet           string
	createdNetworkSubnets          []string
	createdContainerCfg            *runtimecontracts.ContainerConfig
	createdContainerCfgs           []*runtimecontracts.ContainerConfig
	removedContainerID             string
	removedContainerIDs            []string
	removedNetworkID               string
	removedNetworkIDs              []string
	appliedACLRules                []runtimecontracts.InstanceRuntimeACLRule
	removedACLRules                []runtimecontracts.InstanceRuntimeACLRule
	connectedNetworks              map[string][]string
	writtenFiles                   map[string]map[string]string
	imageSize                      int64
	imageInspectErr                error
	removedImageRef                string
	managedContainerStats          []runtimeports.ManagedContainerStat
	managedContainerStates         map[string]*runtimeports.ManagedContainerState
	inspectContainerNetworkIPsFunc func(containerID string, engine *fakeRuntimeEngine) map[string]string
	stopContainerFn                func(ctx context.Context, containerID string, timeout time.Duration) error
	removeContainerFn              func(ctx context.Context, containerID string, force bool) error
	removeNetworkFn                func(ctx context.Context, networkID string) error
	removeACLRulesFn               func(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error
}

func (f *fakeRuntimeEngine) CreateNetwork(_ context.Context, name string, labels map[string]string, _ bool, allowExisting bool, subnet string) (string, error) {
	f.createdNetworkName = name
	f.createdNetworkNames = append(f.createdNetworkNames, name)
	f.createdNetworkAllowExisting = allowExisting
	f.createdNetworkAllowExistingSeq = append(f.createdNetworkAllowExistingSeq, allowExisting)
	f.createdNetworkLabel = labels
	f.createdNetworkSubnet = subnet
	f.createdNetworkSubnets = append(f.createdNetworkSubnets, subnet)
	if len(f.createNetworkErrs) > 0 {
		err := f.createNetworkErrs[0]
		f.createNetworkErrs = f.createNetworkErrs[1:]
		if err != nil {
			return "", err
		}
	}
	if len(f.networkIDs) > 0 {
		networkID := f.networkIDs[0]
		f.networkIDs = f.networkIDs[1:]
		return networkID, nil
	}
	return f.networkID, nil
}

func (f *fakeRuntimeEngine) ListNetworkSubnets(_ context.Context) ([]string, error) {
	f.listNetworkSubnetsCalls++
	if f.listNetworkSubnetsErr != nil {
		return nil, f.listNetworkSubnetsErr
	}
	return append([]string(nil), f.listNetworkSubnets...), nil
}

func (f *fakeRuntimeEngine) CreateContainer(_ context.Context, cfg *runtimecontracts.ContainerConfig) (string, error) {
	f.createdContainerCfg = cfg
	f.createdContainerCfgs = append(f.createdContainerCfgs, cfg)
	if len(f.containerIDs) > 0 {
		containerID := f.containerIDs[0]
		f.containerIDs = f.containerIDs[1:]
		return containerID, nil
	}
	return f.containerID, nil
}

func (f *fakeRuntimeEngine) ResolveServicePort(_ context.Context, _ string, preferredPort int) (int, error) {
	if f.resolveServicePortErr != nil {
		return 0, f.resolveServicePortErr
	}
	if f.resolvedServicePort > 0 {
		return f.resolvedServicePort, nil
	}
	return preferredPort, nil
}

func (f *fakeRuntimeEngine) InspectImageSize(_ context.Context, _ string) (int64, error) {
	if f.imageInspectErr != nil {
		return 0, f.imageInspectErr
	}
	return f.imageSize, nil
}

func (f *fakeRuntimeEngine) RemoveImage(_ context.Context, imageRef string) error {
	f.removedImageRef = imageRef
	return nil
}

func (f *fakeRuntimeEngine) ListManagedContainerStats(_ context.Context) ([]runtimeports.ManagedContainerStat, error) {
	return append([]runtimeports.ManagedContainerStat(nil), f.managedContainerStats...), nil
}

func (f *fakeRuntimeEngine) ConnectContainerToNetwork(_ context.Context, containerID, networkName string) error {
	if f.connectedNetworks == nil {
		f.connectedNetworks = make(map[string][]string)
	}
	f.connectedNetworks[containerID] = append(f.connectedNetworks[containerID], networkName)
	return nil
}

func (f *fakeRuntimeEngine) InspectContainerNetworkIPs(_ context.Context, containerID string) (map[string]string, error) {
	if f.inspectContainerNetworkIPsFunc == nil {
		return nil, nil
	}
	return f.inspectContainerNetworkIPsFunc(containerID, f), nil
}

func (f *fakeRuntimeEngine) StartContainer(_ context.Context, _ string) error {
	return f.startErr
}

func (f *fakeRuntimeEngine) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	if f.stopContainerFn != nil {
		return f.stopContainerFn(ctx, containerID, timeout)
	}
	return nil
}

func (f *fakeRuntimeEngine) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	f.removedContainerID = containerID
	f.removedContainerIDs = append(f.removedContainerIDs, containerID)
	if f.removeContainerFn != nil {
		return f.removeContainerFn(ctx, containerID, force)
	}
	return f.removeContainerErr
}

func (f *fakeRuntimeEngine) RemoveNetwork(ctx context.Context, networkID string) error {
	f.removedNetworkID = networkID
	f.removedNetworkIDs = append(f.removedNetworkIDs, networkID)
	if f.removeNetworkFn != nil {
		return f.removeNetworkFn(ctx, networkID)
	}
	return f.removeNetworkErr
}

func (f *fakeRuntimeEngine) ApplyACLRules(_ context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if f.applyACLErr != nil {
		return f.applyACLErr
	}
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) RemoveACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if f.removeACLRulesFn != nil {
		return f.removeACLRulesFn(ctx, rules)
	}
	f.removedACLRules = append(f.removedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) WriteFileToContainer(_ context.Context, containerID, filePath string, content []byte) error {
	if f.writtenFiles == nil {
		f.writtenFiles = make(map[string]map[string]string)
	}
	if f.writtenFiles[containerID] == nil {
		f.writtenFiles[containerID] = make(map[string]string)
	}
	f.writtenFiles[containerID][filePath] = string(content)
	return nil
}

func (f *fakeRuntimeEngine) ListManagedContainers(_ context.Context) ([]runtimeports.ManagedContainer, error) {
	return nil, nil
}

func (f *fakeRuntimeEngine) InspectManagedContainer(_ context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if f.managedContainerStates == nil {
		return &runtimeports.ManagedContainerState{ID: containerID, Exists: true, Running: true, Status: "running"}, nil
	}
	if state, exists := f.managedContainerStates[containerID]; exists {
		return state, nil
	}
	return &runtimeports.ManagedContainerState{ID: containerID, Exists: false}, nil
}

func seedInstance(t *testing.T, db *gorm.DB, instance *instanceentity.Instance) {
	t.Helper()

	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
}

func seedAWDDefenseWorkspace(t *testing.T, db *gorm.DB, workspace *runtimeentity.AWDDefenseWorkspace) {
	t.Helper()

	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("seed awd defense workspace: %v", err)
	}
}

func seedUser(t *testing.T, db *gorm.DB, user *identitycontracts.User) {
	t.Helper()

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedChallenge(t *testing.T, db *gorm.DB, challenge *runtimeChallengeTestRow) {
	t.Helper()

	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
}
