package composition

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/container_runtime/agentcontracts"
	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	containerruntimeentity "ctf-platform/internal/module/container_runtime/entity"
	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	"ctf-platform/internal/module/container_runtime/infrastructure/agentclient"
	"ctf-platform/internal/module/container_runtime/infrastructure/agentserver"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	runtimeentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
)

func TestRuntimeNodeExecutionRouterRoutesCheckerByNodeID(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	runnerA := &stubRuntimeNodeCheckerRunner{result: contestports.CheckerRunResult{Reason: "node-a"}}
	runnerB := &stubRuntimeNodeCheckerRunner{result: contestports.CheckerRunResult{Reason: "node-b"}}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: &nodeRuntimeClient{checkerRunner: runnerA},
		nodeB.ID: &nodeRuntimeClient{checkerRunner: runnerB},
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	result, err := router.RunChecker(context.Background(), contestports.CheckerRunJob{
		Metadata: contestports.CheckerRunMetadata{NodeID: nodeB.ID},
	})
	if err != nil {
		t.Fatalf("RunChecker() error = %v", err)
	}
	if result.Reason != "node-b" {
		t.Fatalf("expected checker result from node-b, got %+v", result)
	}
	if len(runnerA.jobs) != 0 {
		t.Fatalf("expected node-a checker runner to stay idle, got %d jobs", len(runnerA.jobs))
	}
	if len(runnerB.jobs) != 1 {
		t.Fatalf("expected node-b checker runner to receive 1 job, got %d", len(runnerB.jobs))
	}
}

func TestRuntimeNodeExecutionRouterRejectsExplicitOfflineNodeWhenHealthCheckEnabled(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	cfg.Container.RuntimeNodeHealth.Enabled = true
	cfg.Container.RuntimeNodeHealth.StaleAfter = time.Minute
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	if err := db.Model(nodeB).Updates(map[string]any{
		"health_status": containerruntimeentity.RuntimeNodeHealthOffline,
		"last_seen_at":  time.Now().UTC(),
	}).Error; err != nil {
		t.Fatalf("mark node-b offline: %v", err)
	}

	runnerA := &stubRuntimeNodeCheckerRunner{result: contestports.CheckerRunResult{Reason: "node-a"}}
	runnerB := &stubRuntimeNodeCheckerRunner{result: contestports.CheckerRunResult{Reason: "node-b"}}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: &nodeRuntimeClient{checkerRunner: runnerA},
		nodeB.ID: &nodeRuntimeClient{checkerRunner: runnerB},
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	_, err := router.RunChecker(context.Background(), contestports.CheckerRunJob{
		Metadata: contestports.CheckerRunMetadata{NodeID: nodeB.ID},
	})
	if !errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
		t.Fatalf("RunChecker() error = %v, want ErrRuntimeNodeUnavailable", err)
	}
	if len(runnerB.jobs) != 0 {
		t.Fatalf("expected offline node-b checker runner to stay idle, got %d jobs", len(runnerB.jobs))
	}
}

func TestRuntimeNodeExecutionRouterRoutesExplicitUnschedulableHealthyNodeWhenHealthCheckEnabled(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	cfg.Container.RuntimeNodeHealth.Enabled = true
	cfg.Container.RuntimeNodeHealth.StaleAfter = time.Minute
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)
	now := time.Now().UTC()

	if err := db.Model(nodeA).Updates(map[string]any{
		"health_status": containerruntimeentity.RuntimeNodeHealthReady,
		"last_seen_at":  now,
		"schedulable":   true,
	}).Error; err != nil {
		t.Fatalf("mark node-a ready: %v", err)
	}
	if err := db.Model(nodeB).Updates(map[string]any{
		"health_status": containerruntimeentity.RuntimeNodeHealthReady,
		"last_seen_at":  now,
		"schedulable":   false,
	}).Error; err != nil {
		t.Fatalf("mark node-b cordoned: %v", err)
	}

	runnerA := &stubRuntimeNodeCheckerRunner{result: contestports.CheckerRunResult{Reason: "node-a"}}
	runnerB := &stubRuntimeNodeCheckerRunner{result: contestports.CheckerRunResult{Reason: "node-b"}}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: &nodeRuntimeClient{checkerRunner: runnerA},
		nodeB.ID: &nodeRuntimeClient{checkerRunner: runnerB},
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	defaultResult, err := router.RunChecker(context.Background(), contestports.CheckerRunJob{})
	if err != nil {
		t.Fatalf("default RunChecker() error = %v", err)
	}
	if defaultResult.Reason != "node-a" {
		t.Fatalf("expected default checker to skip cordoned node-b and use node-a, got %+v", defaultResult)
	}

	explicitResult, err := router.RunChecker(context.Background(), contestports.CheckerRunJob{
		Metadata: contestports.CheckerRunMetadata{NodeID: nodeB.ID},
	})
	if err != nil {
		t.Fatalf("explicit RunChecker() error = %v", err)
	}
	if explicitResult.Reason != "node-b" {
		t.Fatalf("expected explicit checker result from cordoned node-b, got %+v", explicitResult)
	}
	if len(runnerA.jobs) != 1 {
		t.Fatalf("expected node-a checker runner to receive default job only, got %d", len(runnerA.jobs))
	}
	if len(runnerB.jobs) != 1 {
		t.Fatalf("expected node-b checker runner to receive explicit job only, got %d", len(runnerB.jobs))
	}
}

func TestBuildRuntimeNodeClientRejectsRuntimeAgentNodeNameMismatch(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:          true,
		NodeName:         "node-a",
		Endpoint:         "runtime-agent-a:9443",
		DialTimeout:      time.Second,
		ServerName:       "runtime-agent-a",
		CAFile:           "/etc/ctf/ca.pem",
		CertFile:         "/etc/ctf/client.pem",
		KeyFile:          "/etc/ctf/client-key.pem",
		KeepaliveTime:    30 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
	}
	node := &containerruntimeentity.RuntimeNode{
		ID:          101,
		Name:        "node-a",
		Endpoint:    "runtime-agent-a:9443",
		TLSIdentity: "runtime-agent-a",
	}
	originalDial := dialRuntimeAgent
	dialRuntimeAgent = func(ctx context.Context, _ config.RuntimeAgentConfig) (*agentclient.Client, error) {
		return newRuntimeAgentHealthClient(t, ctx, agentserver.ServiceIdentity{
			NodeName: "node-b",
			Hostname: "agent-host-b",
		}), nil
	}
	t.Cleanup(func() {
		dialRuntimeAgent = originalDial
	})

	client, err := buildRuntimeNodeClientFromNode(
		context.Background(),
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		node,
	)
	if err == nil {
		if client != nil {
			_ = client.Close(context.Background())
		}
		t.Fatal("expected runtime agent node identity mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "runtime agent node identity mismatch") ||
		!strings.Contains(err.Error(), "node-a") ||
		!strings.Contains(err.Error(), "node-b") ||
		!strings.Contains(err.Error(), "agent-host-b") {
		t.Fatalf("unexpected mismatch error: %v", err)
	}
}

func TestBuildRuntimeNodeClientTimesOutRuntimeAgentIdentityCheck(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:          true,
		NodeName:         "node-a",
		Endpoint:         "runtime-agent-a:9443",
		DialTimeout:      20 * time.Millisecond,
		ServerName:       "runtime-agent-a",
		CAFile:           "/etc/ctf/ca.pem",
		CertFile:         "/etc/ctf/client.pem",
		KeyFile:          "/etc/ctf/client-key.pem",
		KeepaliveTime:    30 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
	}
	node := &containerruntimeentity.RuntimeNode{
		ID:          101,
		Name:        "node-a",
		Endpoint:    "runtime-agent-a:9443",
		TLSIdentity: "runtime-agent-a",
	}
	originalDial := dialRuntimeAgent
	dialRuntimeAgent = func(ctx context.Context, _ config.RuntimeAgentConfig) (*agentclient.Client, error) {
		return newRuntimeAgentClient(t, ctx, blockingRuntimeAgentService{}), nil
	}
	t.Cleanup(func() {
		dialRuntimeAgent = originalDial
	})

	done := make(chan error, 1)
	go func() {
		client, err := buildRuntimeNodeClientFromNode(
			context.Background(),
			cfg,
			zap.NewNop(),
			containerruntimeinfra.NewAllocationRepository(db),
			node,
		)
		if client != nil {
			_ = client.Close(context.Background())
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected runtime agent identity timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "check runtime agent node identity") {
			t.Fatalf("unexpected timeout error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runtime agent identity check did not respect configured timeout")
	}
}

func TestRuntimeNodeExecutionRouterRoutesContainerFileWritesByWorkspaceContainerNodeID(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	if err := db.AutoMigrate(&instanceentity.Instance{}, &runtimeentity.AWDDefenseWorkspace{}); err != nil {
		t.Fatalf("auto migrate router dependencies: %v", err)
	}

	nodeBID := nodeB.ID
	instance := instanceentity.Instance{
		ID:          2001,
		UserID:      3001,
		ContestID:   int64PtrForRouterTest(41),
		TeamID:      int64PtrForRouterTest(51),
		ChallengeID: 61,
		ServiceID:   int64PtrForRouterTest(71),
		NodeID:      &nodeBID,
		ContainerID: "primary-ctr",
		ShareScope:  instanceentity.ShareScopePerUser,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
	workspace := runtimeentity.AWDDefenseWorkspace{
		ContestID:         41,
		TeamID:            51,
		ServiceID:         71,
		InstanceID:        instance.ID,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr",
		SeedSignature:     "seed-v1",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("create defense workspace: %v", err)
	}

	executorA := &stubRuntimeNodeHostExecutor{}
	executorB := &stubRuntimeNodeHostExecutor{}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	if err := router.WriteFileToContainer(context.Background(), "workspace-ctr", "/opt/checker/flag.txt", []byte("payload")); err != nil {
		t.Fatalf("WriteFileToContainer() error = %v", err)
	}
	if len(executorA.writeCalls) != 0 {
		t.Fatalf("expected node-a executor to stay idle, got %+v", executorA.writeCalls)
	}
	if len(executorB.writeCalls) != 1 {
		t.Fatalf("expected node-b executor to receive 1 write, got %+v", executorB.writeCalls)
	}
	call := executorB.writeCalls[0]
	if call.containerID != "workspace-ctr" || call.filePath != "/opt/checker/flag.txt" || string(call.content) != "payload" {
		t.Fatalf("unexpected write call: %+v", call)
	}
}

func TestRuntimeNodeExecutionRouterRoutesInteractiveExecByWorkspaceContainerNodeID(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	if err := db.AutoMigrate(&instanceentity.Instance{}, &runtimeentity.AWDDefenseWorkspace{}); err != nil {
		t.Fatalf("auto migrate interactive exec dependencies: %v", err)
	}

	nodeBID := nodeB.ID
	instance := instanceentity.Instance{
		ID:          2051,
		UserID:      3051,
		ContestID:   int64PtrForRouterTest(43),
		TeamID:      int64PtrForRouterTest(53),
		ChallengeID: 63,
		ServiceID:   int64PtrForRouterTest(73),
		NodeID:      &nodeBID,
		ContainerID: "primary-ctr",
		ShareScope:  instanceentity.ShareScopePerUser,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
	workspace := runtimeentity.AWDDefenseWorkspace{
		ContestID:         43,
		TeamID:            53,
		ServiceID:         73,
		InstanceID:        instance.ID,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-interactive-ctr",
		SeedSignature:     "seed-v3",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("create defense workspace: %v", err)
	}

	executorA := &stubRuntimeNodeHostExecutor{}
	executorB := &stubRuntimeNodeHostExecutor{}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	if err := router.ExecContainerInteractive(context.Background(), "workspace-interactive-ctr", []string{"sh", "-lc", "pwd"}, nil, io.Discard); err != nil {
		t.Fatalf("ExecContainerInteractive() error = %v", err)
	}
	if len(executorA.interactiveCalls) != 0 {
		t.Fatalf("expected node-a executor to stay idle, got %+v", executorA.interactiveCalls)
	}
	if len(executorB.interactiveCalls) != 1 {
		t.Fatalf("expected node-b executor to receive 1 interactive call, got %+v", executorB.interactiveCalls)
	}
	call := executorB.interactiveCalls[0]
	if call.containerID != "workspace-interactive-ctr" {
		t.Fatalf("unexpected interactive container id: %+v", call)
	}
	if len(call.command) != 3 || call.command[0] != "sh" || call.command[2] != "pwd" {
		t.Fatalf("unexpected interactive command: %+v", call)
	}
}

func TestRuntimeNodeExecutionRouterRoutesCleanupByRuntimeDetailsContainerNodeID(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	if err := db.AutoMigrate(&instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate instances: %v", err)
	}

	nodeBID := nodeB.ID
	storedRuntimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{ContainerID: "runtime-details-ctr"},
		},
	})
	if err != nil {
		t.Fatalf("encode stored runtime details: %v", err)
	}
	storedInstance := instanceentity.Instance{
		ID:             2101,
		UserID:         3101,
		ChallengeID:    4101,
		NodeID:         &nodeBID,
		RuntimeDetails: storedRuntimeDetails,
		ShareScope:     instanceentity.ShareScopePerUser,
		Status:         instanceentity.InstanceStatusRunning,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&storedInstance).Error; err != nil {
		t.Fatalf("create stored instance: %v", err)
	}

	executorA := &stubRuntimeNodeHostExecutor{}
	executorB := &stubRuntimeNodeHostExecutor{}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	cleanupPayloadDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{ContainerID: "runtime-details-ctr"},
		},
	})
	if err != nil {
		t.Fatalf("encode cleanup runtime details: %v", err)
	}
	if err := router.CleanupRuntime(context.Background(), runtimeCleanupTargetFromInstance(&instanceentity.Instance{RuntimeDetails: cleanupPayloadDetails})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	if len(executorA.removedContainers) != 0 {
		t.Fatalf("expected node-a executor to stay idle, got %+v", executorA.removedContainers)
	}
	if len(executorB.removedContainers) != 1 || executorB.removedContainers[0] != "runtime-details-ctr" {
		t.Fatalf("expected node-b executor to remove runtime-details-ctr, got %+v", executorB.removedContainers)
	}
}

func TestRuntimeNodeExecutionRouterRoutesCleanupByWorkspaceContainerIDWithoutNodeID(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	if err := db.AutoMigrate(&instanceentity.Instance{}, &runtimeentity.AWDDefenseWorkspace{}); err != nil {
		t.Fatalf("auto migrate cleanup dependencies: %v", err)
	}

	nodeBID := nodeB.ID
	storedInstance := instanceentity.Instance{
		ID:          2201,
		UserID:      3201,
		ContestID:   int64PtrForRouterTest(42),
		TeamID:      int64PtrForRouterTest(52),
		ServiceID:   int64PtrForRouterTest(72),
		ChallengeID: 62,
		NodeID:      &nodeBID,
		ContainerID: "primary-ctr",
		ShareScope:  instanceentity.ShareScopePerUser,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&storedInstance).Error; err != nil {
		t.Fatalf("create stored instance: %v", err)
	}
	workspace := runtimeentity.AWDDefenseWorkspace{
		ContestID:         42,
		TeamID:            52,
		ServiceID:         72,
		InstanceID:        storedInstance.ID,
		WorkspaceRevision: 2,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-cleanup-ctr",
		SeedSignature:     "seed-v2",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("create defense workspace: %v", err)
	}

	executorA := &stubRuntimeNodeHostExecutor{}
	executorB := &stubRuntimeNodeHostExecutor{}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	if err := router.CleanupRuntime(context.Background(), runtimeCleanupTargetFromInstance(&instanceentity.Instance{ContainerID: "workspace-cleanup-ctr"})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	if len(executorA.removedContainers) != 0 {
		t.Fatalf("expected node-a executor to stay idle, got %+v", executorA.removedContainers)
	}
	if len(executorB.removedContainers) != 1 || executorB.removedContainers[0] != "workspace-cleanup-ctr" {
		t.Fatalf("expected node-b executor to remove workspace-cleanup-ctr, got %+v", executorB.removedContainers)
	}
}

func TestRuntimeNodeExecutionRouterRoutesRemoveContainerByInventoryCache(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	executorA := &stubRuntimeNodeHostExecutor{
		listManagedContainersResult: []runtimecontracts.ManagedContainer{{ID: "node-a-ctr"}},
	}
	executorB := &stubRuntimeNodeHostExecutor{
		listManagedContainersResult: []runtimecontracts.ManagedContainer{{ID: "orphan-ctr"}},
	}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	containers, err := router.ListManagedContainers(context.Background())
	if err != nil {
		t.Fatalf("ListManagedContainers() error = %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 managed containers, got %+v", containers)
	}

	if err := router.RemoveContainer(context.Background(), "orphan-ctr"); err != nil {
		t.Fatalf("RemoveContainer() error = %v", err)
	}
	if len(executorA.removedContainers) != 0 {
		t.Fatalf("expected node-a executor to stay idle, got %+v", executorA.removedContainers)
	}
	if len(executorB.removedContainers) != 1 || executorB.removedContainers[0] != "orphan-ctr" {
		t.Fatalf("expected node-b executor to remove orphan-ctr, got %+v", executorB.removedContainers)
	}
}

func TestRuntimeNodeExecutionRouterListsManagedContainersFromUnschedulableNode(t *testing.T) {
	cfg, db, _ := newRootTestDependencies(t)
	nodeA, nodeB := seedRuntimeRouterNodes(t, db)

	if err := db.Model(nodeB).Update("schedulable", false).Error; err != nil {
		t.Fatalf("cordon node-b: %v", err)
	}

	executorA := &stubRuntimeNodeHostExecutor{
		listManagedContainersResult: []runtimecontracts.ManagedContainer{{ID: "node-a-ctr"}},
	}
	executorB := &stubRuntimeNodeHostExecutor{
		listManagedContainersResult: []runtimecontracts.ManagedContainer{{ID: "node-b-existing-ctr"}},
	}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		containerruntimeinfra.NewAllocationRepository(db),
		newRuntimeNodeTestContainerIndex(db),
		containerruntimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	containers, err := router.ListManagedContainers(context.Background())
	if err != nil {
		t.Fatalf("ListManagedContainers() error = %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected inventory from schedulable and cordoned nodes, got %+v", containers)
	}
	nodeID, err := router.resolveNodeIDForContainer(context.Background(), "node-b-existing-ctr")
	if err != nil {
		t.Fatalf("resolve node-b inventory cache: %v", err)
	}
	if nodeID != nodeB.ID {
		t.Fatalf("cached node id = %d, want %d", nodeID, nodeB.ID)
	}
}

func seedRuntimeRouterNodes(t *testing.T, db *gorm.DB) (*containerruntimeentity.RuntimeNode, *containerruntimeentity.RuntimeNode) {
	t.Helper()

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}); err != nil {
		t.Fatalf("auto migrate runtime nodes: %v", err)
	}

	nodeA := &containerruntimeentity.RuntimeNode{
		Name:             "node-a",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(nodeA).Error; err != nil {
		t.Fatalf("create node-a: %v", err)
	}

	nodeB := &containerruntimeentity.RuntimeNode{
		Name:             "node-b",
		Endpoint:         "grpc://runtime-agent-b",
		TLSIdentity:      "runtime-agent-b",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(nodeB).Error; err != nil {
		t.Fatalf("create node-b: %v", err)
	}
	return nodeA, nodeB
}

func newRuntimeNodeTestContainerIndex(db *gorm.DB) runtimeNodeStateRepository {
	return newCompositeRuntimeNodeContainerIndex(
		newInstanceRuntimeInventoryProvider(instanceinfra.NewContainerInventoryRepository(db)),
		contestinfra.NewAWDContainerInventoryRepository(db),
	)
}

func overrideRuntimeNodeClientBuilder(t *testing.T, clients map[int64]runtimeNodeClient) {
	t.Helper()

	original := buildRuntimeNodeClient
	buildRuntimeNodeClient = func(_ context.Context, _ *config.Config, _ *zap.Logger, _ runtimeNodeAllocationRepository, node *containerruntimeentity.RuntimeNode) (runtimeNodeClient, error) {
		if node == nil {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		client, ok := clients[node.ID]
		if !ok || client == nil {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		return client, nil
	}
	t.Cleanup(func() {
		buildRuntimeNodeClient = original
	})
}

func newRuntimeAgentHealthClient(t *testing.T, ctx context.Context, identity agentserver.ServiceIdentity) *agentclient.Client {
	t.Helper()
	return newRuntimeAgentClient(t, ctx, agentserver.NewService(nil, nil, identity))
}

func newRuntimeAgentClient(t *testing.T, ctx context.Context, service agentcontracts.RuntimeAgentService) *agentclient.Client {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.ForceServerCodec(agentcontracts.JSONCodec()))
	agentcontracts.RegisterRuntimeAgentService(server, service)
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("runtime agent health bufconn stopped: %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(agentcontracts.JSONCodec())),
	)
	if err != nil {
		t.Fatalf("dial runtime agent health bufconn: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
		server.Stop()
		_ = listener.Close()
	})
	return agentclient.New(conn)
}

type blockingRuntimeAgentService struct {
	agentcontracts.RuntimeAgentService
}

func (blockingRuntimeAgentService) Health(ctx context.Context, _ *agentcontracts.HealthRequest) (*agentcontracts.HealthResponse, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func newStubNodeRuntimeClient(executor runtimeports.RuntimeHostExecutor, runner contestports.CheckerRunner) *nodeRuntimeClient {
	client := &nodeRuntimeClient{
		executor:      executor,
		checkerRunner: runner,
	}
	if executor != nil {
		client.cleaner = runtimecmd.NewRuntimeCleanupService(executor, nil, zap.NewNop())
	}
	return client
}

func int64PtrForRouterTest(value int64) *int64 {
	return &value
}

type stubRuntimeNodeCheckerRunner struct {
	jobs   []contestports.CheckerRunJob
	result contestports.CheckerRunResult
	err    error
}

func (s *stubRuntimeNodeCheckerRunner) RunChecker(_ context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	s.jobs = append(s.jobs, job)
	return s.result, s.err
}

type runtimeNodeWriteCall struct {
	containerID string
	filePath    string
	content     []byte
}

type runtimeNodeInteractiveCall struct {
	containerID string
	command     []string
}

type stubRuntimeNodeHostExecutor struct {
	writeCalls                  []runtimeNodeWriteCall
	interactiveCalls            []runtimeNodeInteractiveCall
	listManagedContainersResult []runtimecontracts.ManagedContainer
	removedContainers           []string
	appliedACLCalls             []stubRuntimeNodeACLCall
	removedACLRulesCalls        [][]runtimecontracts.InstanceRuntimeACLRule
	removeACLRulesErr           error
	startedContainers           []string
}

type stubRuntimeNodeACLCall struct {
	handle *runtimecontracts.InstanceRuntimeACLHandle
	rules  []runtimecontracts.InstanceRuntimeACLRule
}

func (s *stubRuntimeNodeHostExecutor) CreateNetwork(context.Context, string, map[string]string, bool, bool, string) (string, error) {
	return "", nil
}

func (s *stubRuntimeNodeHostExecutor) ListNetworkSubnets(context.Context) ([]string, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) CreateContainer(context.Context, *runtimecontracts.ContainerConfig) (string, error) {
	return "", nil
}

func (s *stubRuntimeNodeHostExecutor) ResolveServicePort(context.Context, string, int) (int, error) {
	return 0, nil
}

func (s *stubRuntimeNodeHostExecutor) ConnectContainerToNetwork(context.Context, string, string) error {
	return nil
}

func (s *stubRuntimeNodeHostExecutor) InspectContainerNetworkIPs(context.Context, string) (map[string]string, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) StartContainer(_ context.Context, containerID string) error {
	s.startedContainers = append(s.startedContainers, containerID)
	return nil
}

func (s *stubRuntimeNodeHostExecutor) StopContainer(context.Context, string, time.Duration) error {
	return nil
}

func (s *stubRuntimeNodeHostExecutor) RemoveContainer(_ context.Context, containerID string, _ bool) error {
	s.removedContainers = append(s.removedContainers, containerID)
	return nil
}

func (s *stubRuntimeNodeHostExecutor) RemoveNetwork(context.Context, string) error {
	return nil
}

func (s *stubRuntimeNodeHostExecutor) ApplyACLRules(context.Context, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (s *stubRuntimeNodeHostExecutor) ApplyACL(_ context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	var copiedHandle *runtimecontracts.InstanceRuntimeACLHandle
	if handle != nil {
		handleCopy := *handle
		copiedHandle = &handleCopy
	}
	s.appliedACLCalls = append(s.appliedACLCalls, stubRuntimeNodeACLCall{
		handle: copiedHandle,
		rules:  append([]runtimecontracts.InstanceRuntimeACLRule(nil), rules...),
	})
	return nil
}

func (s *stubRuntimeNodeHostExecutor) RemoveACLRules(_ context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	s.removedACLRulesCalls = append(s.removedACLRulesCalls, append([]runtimecontracts.InstanceRuntimeACLRule(nil), rules...))
	return s.removeACLRulesErr
}

func (s *stubRuntimeNodeHostExecutor) RemoveACL(context.Context, *runtimecontracts.InstanceRuntimeACLHandle) error {
	return nil
}

func (s *stubRuntimeNodeHostExecutor) WriteFileToContainer(_ context.Context, containerID, filePath string, content []byte) error {
	s.writeCalls = append(s.writeCalls, runtimeNodeWriteCall{
		containerID: containerID,
		filePath:    filePath,
		content:     append([]byte(nil), content...),
	})
	return nil
}

func (s *stubRuntimeNodeHostExecutor) ReadFileFromContainer(context.Context, string, string, int64) ([]byte, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimecontracts.ContainerDirectoryEntry, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) InspectImageSize(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *stubRuntimeNodeHostExecutor) RemoveImage(context.Context, string) error {
	return nil
}

func (s *stubRuntimeNodeHostExecutor) ListManagedContainers(context.Context) ([]runtimecontracts.ManagedContainer, error) {
	return append([]runtimecontracts.ManagedContainer(nil), s.listManagedContainersResult...), nil
}

func (s *stubRuntimeNodeHostExecutor) InspectManagedContainer(context.Context, string) (*runtimecontracts.ManagedContainerState, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) ListManagedContainerStats(context.Context) ([]runtimecontracts.ManagedContainerStat, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) ExecContainerInteractive(_ context.Context, containerID string, command []string, _ io.Reader, _ io.Writer) error {
	s.interactiveCalls = append(s.interactiveCalls, runtimeNodeInteractiveCall{
		containerID: containerID,
		command:     append([]string(nil), command...),
	})
	return nil
}
