package runtime_test

import (
	"context"
	"errors"
	"testing"

	"ctf-platform/internal/config"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

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
	if got := engine.createdContainerCfg.Labels[runtimecontracts.ComposeProjectLabelKey]; got != runtimecontracts.ProjectLabelValue {
		t.Fatalf("expected compose project label %q, got %q", runtimecontracts.ProjectLabelValue, got)
	}
	if got := engine.createdContainerCfg.Labels[runtimecontracts.ComposeServiceLabelKey]; got != runtimecontracts.ComposeServiceJeopardy {
		t.Fatalf("expected jeopardy compose service label, got %q", got)
	}
	if got := engine.createdNetworkLabel[runtimecontracts.ComposeServiceLabelKey]; got != runtimecontracts.ComposeServiceJeopardy {
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
