package composition

import (
	"context"
	"io"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
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
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
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
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
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
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
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
	if err := router.CleanupRuntime(context.Background(), &instanceentity.Instance{RuntimeDetails: cleanupPayloadDetails}); err != nil {
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
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
		"",
	)

	if err := router.CleanupRuntime(context.Background(), &instanceentity.Instance{ContainerID: "workspace-cleanup-ctr"}); err != nil {
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
		listManagedContainersResult: []runtimeports.ManagedContainer{{ID: "node-a-ctr"}},
	}
	executorB := &stubRuntimeNodeHostExecutor{
		listManagedContainersResult: []runtimeports.ManagedContainer{{ID: "orphan-ctr"}},
	}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		nodeA.ID: newStubNodeRuntimeClient(executorA, nil),
		nodeB.ID: newStubNodeRuntimeClient(executorB, nil),
	})

	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
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

func seedRuntimeRouterNodes(t *testing.T, db *gorm.DB) (*runtimeentity.RuntimeNode, *runtimeentity.RuntimeNode) {
	t.Helper()

	if err := db.AutoMigrate(&runtimeentity.RuntimeNode{}); err != nil {
		t.Fatalf("auto migrate runtime nodes: %v", err)
	}

	nodeA := &runtimeentity.RuntimeNode{
		Name:             "node-a",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     runtimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(nodeA).Error; err != nil {
		t.Fatalf("create node-a: %v", err)
	}

	nodeB := &runtimeentity.RuntimeNode{
		Name:             "node-b",
		Endpoint:         "grpc://runtime-agent-b",
		TLSIdentity:      "runtime-agent-b",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     runtimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(nodeB).Error; err != nil {
		t.Fatalf("create node-b: %v", err)
	}
	return nodeA, nodeB
}

func overrideRuntimeNodeClientBuilder(t *testing.T, clients map[int64]runtimeNodeClient) {
	t.Helper()

	original := buildRuntimeNodeClient
	buildRuntimeNodeClient = func(_ context.Context, _ *config.Config, _ *zap.Logger, _ *runtimeinfra.Repository, node *runtimeentity.RuntimeNode) (runtimeNodeClient, error) {
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

type stubRuntimeNodeHostExecutor struct {
	writeCalls                  []runtimeNodeWriteCall
	listManagedContainersResult []runtimeports.ManagedContainer
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

func (s *stubRuntimeNodeHostExecutor) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimeports.ContainerDirectoryEntry, error) {
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

func (s *stubRuntimeNodeHostExecutor) ListManagedContainers(context.Context) ([]runtimeports.ManagedContainer, error) {
	return append([]runtimeports.ManagedContainer(nil), s.listManagedContainersResult...), nil
}

func (s *stubRuntimeNodeHostExecutor) InspectManagedContainer(context.Context, string) (*runtimeports.ManagedContainerState, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) ListManagedContainerStats(context.Context) ([]runtimeports.ManagedContainerStat, error) {
	return nil, nil
}

func (s *stubRuntimeNodeHostExecutor) ExecContainerInteractive(context.Context, string, []string, io.Reader, io.Writer) error {
	return nil
}
