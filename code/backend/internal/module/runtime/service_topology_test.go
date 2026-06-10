package runtime_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	containerruntimeentity "ctf-platform/internal/module/container_runtime/entity"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
	if err := repo.db.Model(&containerruntimeentity.PortAllocation{}).Count(&count).Error; err != nil {
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
	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		ContainerName: preferredName,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
	if got := engine.createdContainerCfg.Labels[runtimecontracts.ComposeServiceLabelKey]; got != runtimecontracts.ComposeServiceAWD {
		t.Fatalf("expected awd compose service label, got %q", got)
	}
	if got := engine.createdNetworkLabel[runtimecontracts.ComposeServiceLabelKey]; got != runtimecontracts.ComposeServiceAWD {
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

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		ContainerName:              "ctf-workspace-workspace-c8-t15-s21-r2",
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
	if got := engine.createdContainerCfg.Labels[runtimecontracts.ComposeServiceLabelKey]; got != runtimecontracts.ComposeServiceAWD {
		t.Fatalf("expected awd compose service label, got %q", got)
	}
	if got := engine.createdNetworkLabel[runtimecontracts.ComposeServiceLabelKey]; got != runtimecontracts.ComposeServiceAWD {
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

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
		RuntimeDetails: `{"containers":[{"container_id":"web-ctr"},{"container_id":"db-ctr"}],"acl":{"chain":"CTF-INS-1"},"acl_rules":[{"comment":"ctf:acl:test","source_ip":"172.30.0.2","target_ip":"172.30.0.3","action":"allow","protocol":"tcp","ports":[3306]}]}`,
		Status:         instanceentity.InstanceStatusRunning,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&containerruntimeentity.PortAllocation{Port: 30001, InstanceID: &instance.ID}).Error; err != nil {
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
	if engine.removedACLHandle == nil || engine.removedACLHandle.Chain != "CTF-INS-1" {
		t.Fatalf("expected acl handle to be removed, got %+v", engine.removedACLHandle)
	}
	if len(engine.removedACLRules) != 0 {
		t.Fatalf("expected no legacy acl rule removal, got %+v", engine.removedACLRules)
	}

	if updated.Status != instanceentity.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %+v", updated)
	}
	if updated.HostPort != 0 || updated.ContainerID != "" || updated.NetworkID != "" || updated.RuntimeDetails != "" || updated.AccessURL != "" {
		t.Fatalf("expected stopped instance runtime fields to be cleared, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30001).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
}

func TestServiceCleanExpiredInstancesKeepsRunningStateWhenRuntimeCleanupFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now().UTC()
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
	if err := repo.db.Create(&containerruntimeentity.PortAllocation{
		Port:       30002,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	engine := &fakeRuntimeEngine{removeContainerErr: errors.New("remove failed")}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)
	service := instancecmd.NewInstanceMaintenanceService(newRuntimeTestMaintenanceRepository(repo), nil, newRuntimeTestCleanerAdapter(cleanupService), &config.ContainerConfig{
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
	if err := repo.db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30002).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected port allocation to remain reserved, count=%d", count)
	}
}

func TestServiceCleanExpiredInstancesMarksExpiredWhenContainerAlreadyRemoved(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now().UTC()
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
	if err := repo.db.Create(&containerruntimeentity.PortAllocation{
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
	service := instancecmd.NewInstanceMaintenanceService(newRuntimeTestMaintenanceRepository(repo), nil, newRuntimeTestCleanerAdapter(cleanupService), &config.ContainerConfig{
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
	if err := repo.db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30003).Count(&count).Error; err != nil {
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
	if err := repo.db.Create(&containerruntimeentity.PortAllocation{
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
	if err := repo.db.Model(&containerruntimeentity.PortAllocation{}).Where("port = ?", 30004).Count(&count).Error; err != nil {
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: "public"},
			{Key: "backend", Internal: true},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 4242,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
			{Key: "extra"},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
		assertNonNegativeLogDuration(t, ctxMap)
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 7001,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 7002,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 7003,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Subnet: "10.10.1.0/24"},
			{Key: "backend"},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 7005,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Subnet: "10.10.10.0/24"},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
	if err := repo.db.Create(&containerruntimeentity.NetworkAllocation{
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: instanceID,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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

	var allocation containerruntimeentity.NetworkAllocation
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

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 5252,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
	assertNonNegativeLogDuration(t, ctxMap)
	if _, exists := ctxMap["error"]; !exists {
		t.Fatalf("expected failure log to include error field, got %+v", ctxMap)
	}
}

func assertNonNegativeLogDuration(t *testing.T, ctxMap map[string]interface{}) {
	t.Helper()

	duration, ok := ctxMap["duration"].(time.Duration)
	if !ok {
		t.Fatalf("expected duration field in stage log, got %+v", ctxMap)
	}
	if duration < 0 {
		t.Fatalf("expected non-negative duration, got %s in %+v", duration, ctxMap)
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

	result, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 4242,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
	if engine.appliedACLHandle == nil || engine.appliedACLHandle.Chain != "CTF-INS-4242" {
		t.Fatalf("expected acl handle CTF-INS-4242, got %+v", engine.appliedACLHandle)
	}
	if len(engine.appliedACLRules) != 2 {
		t.Fatalf("expected 2 acl rules, got %+v", engine.appliedACLRules)
	}
	if result.RuntimeDetails.ACL == nil || result.RuntimeDetails.ACL.Chain != "CTF-INS-4242" {
		t.Fatalf("expected runtime acl handle, got %+v", result.RuntimeDetails.ACL)
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

func TestServiceCreateTopologyRejectsACLWithoutOwnerInstanceID(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-no-owner",
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

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{runtimecontracts.TopologyDefaultNetworkKey}},
		},
		Policies: []runtimecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: runtimecontracts.TopologyPolicyActionAllow, Protocol: runtimecontracts.TopologyPolicyProtocolTCP, Ports: []int{3306}},
		},
	})
	if err == nil {
		t.Fatal("expected CreateTopology() to fail without owner instance id")
	}
	if !strings.Contains(err.Error(), "owner instance id") {
		t.Fatalf("expected owner instance id error, got %v", err)
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

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		OwnerInstanceID: 1,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
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
