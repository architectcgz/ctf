package runtime_test

import (
	"context"
	"testing"
	"time"

	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
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

	instances, err := repo.ListStoppingInstances(context.Background(), cutoff, 0)
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

func TestRepositoryListStoppingInstancesAppliesLimit(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	updatedAt := now.Add(-10 * time.Minute)

	for i := int64(0); i < 4; i++ {
		seedInstance(t, repo.db, &instanceentity.Instance{
			ID:          401 + i,
			UserID:      1,
			ChallengeID: 301 + i,
			Status:      instanceentity.InstanceStatusStopping,
			ExpiresAt:   now.Add(time.Hour),
			CreatedAt:   updatedAt,
			UpdatedAt:   updatedAt.Add(time.Duration(i) * time.Second),
		})
	}

	instances, err := repo.ListStoppingInstances(context.Background(), time.Time{}, 2)
	if err != nil {
		t.Fatalf("ListStoppingInstances() error = %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 stopping instances, got %d", len(instances))
	}
	if instances[0].ID != 401 || instances[1].ID != 402 {
		t.Fatalf("expected oldest stopping instances first, got ids=%d,%d", instances[0].ID, instances[1].ID)
	}
}
