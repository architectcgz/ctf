package runtime_test

import (
	"context"
	"testing"
	"time"

	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

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

	if err := service.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{RuntimeDetails: runtimeDetails})); err != nil {
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

	if err := service.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{ID: instanceID, RuntimeDetails: runtimeDetails})); err != nil {
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

	if err := service.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{ID: instanceID, RuntimeDetails: runtimeDetails})); err != nil {
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

	if err := service.CleanupRuntime(context.Background(), runtimeCleanupTarget(&instanceentity.Instance{ID: ownerInstanceID, RuntimeDetails: runtimeDetails})); err != nil {
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
