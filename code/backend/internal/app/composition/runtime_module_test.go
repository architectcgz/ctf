package composition

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"ctf-platform/internal/authctx"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	containerruntimeentity "ctf-platform/internal/module/container_runtime/entity"
	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
)

func TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container = config.ContainerConfig{
		DefaultExposedPort: 80,
		PortRangeStart:     35000,
		PortRangeEnd:       35010,
		PublicHost:         "127.0.0.1",
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}
	if err := db.AutoMigrate(&containerruntimeentity.NetworkAllocation{}); err != nil {
		t.Fatalf("auto migrate runtime network allocation: %v", err)
	}
	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}); err != nil {
		t.Fatalf("auto migrate runtime nodes: %v", err)
	}

	executor := buildRuntimeHostExecutor(root)
	service := runtimecmd.NewProvisioningService(containerruntimeinfra.NewAllocationRepository(db), executor, &cfg.Container, zap.NewNop())

	containerID, networkID, hostPort, _, err := service.CreateContainer(context.Background(), "ctf/test:v1", nil, 35001)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	if containerID == "" {
		t.Fatal("expected non-empty container id")
	}
	if networkID == "" {
		t.Fatal("expected non-empty network id")
	}
	if hostPort != 35001 {
		t.Fatalf("expected host port 35001, got %d", hostPort)
	}

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://127.0.0.1:35001")
	if err != nil {
		t.Fatalf("expected runtime access url to be reachable, got error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected runtime probe status 200, got %d", resp.StatusCode)
	}

	cleanup := runtimecmd.NewRuntimeCleanupService(executor, nil, zap.NewNop())
	if err := cleanup.RemoveContainer(context.Background(), containerID); err != nil {
		t.Fatalf("RemoveContainer() error = %v", err)
	}
	if executor != nil && networkID != "" {
		if err := executor.RemoveNetwork(context.Background(), networkID); err != nil {
			t.Fatalf("RemoveNetwork() error = %v", err)
		}
	}
}

func TestRuntimePublishedAccessHostAllowsNilConfig(t *testing.T) {
	t.Parallel()

	if got := runtimePublishedAccessHost(nil); got != "" {
		t.Fatalf("expected empty runtime access host for nil config, got %q", got)
	}
}

func TestRuntimeHTTPServiceAdapterReturnsSSHAccessWithoutProfile(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		stubRuntimeHTTPProxyTickets{ticket: "ticket-secret", expiresAt: expiresAt},
		0,
		0,
		true,
		"ssh.ctf.local",
		2222,
	)

	resp, err := adapter.IssueAWDDefenseSSHTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "student",
		Role:     "student",
	}, 5, 12)
	if err != nil {
		t.Fatalf("IssueAWDDefenseSSHTicket() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected ssh access response")
	}
	if resp.Host != "ssh.ctf.local" ||
		resp.Port != 2222 ||
		resp.Username != "student+5+12" ||
		resp.Password != "ticket-secret" ||
		resp.Command != "ssh student+5+12@ssh.ctf.local -p 2222" ||
		resp.ExpiresAt != expiresAt.Format(time.RFC3339) {
		t.Fatalf("unexpected ssh access response: %+v", resp)
	}
}

func TestRuntimeHTTPServiceAdapterUsesExternalSSHHostNotBindAddress(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 6, 12, 8, 0, 0, 0, time.UTC)
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		stubRuntimeHTTPProxyTickets{ticket: "ticket-secret", expiresAt: expiresAt},
		0,
		0,
		true,
		"lb.ctf.example",
		30222,
	)

	resp, err := adapter.IssueAWDDefenseSSHTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "student",
		Role:     "student",
	}, 11, 22)
	if err != nil {
		t.Fatalf("IssueAWDDefenseSSHTicket() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected ssh access response")
	}
	if resp.Host != "lb.ctf.example" || resp.Port != 30222 {
		t.Fatalf("expected external lb host/port, got %+v", resp)
	}
	if resp.Command != "ssh student+11+22@lb.ctf.example -p 30222" {
		t.Fatalf("unexpected ssh command: %q", resp.Command)
	}
}

func TestBuildContainerRuntimeModuleFailsWhenRemoteRuntimeAgentDialFails(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	caFile, certFile, keyFile := writeRemoteAgentClientTLSFiles(t)
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:     true,
		Endpoint:    "127.0.0.1:1",
		DialTimeout: 50 * time.Millisecond,
		ServerName:  "runtime-agent.local",
		CAFile:      caFile,
		CertFile:    certFile,
		KeyFile:     keyFile,
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}
	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}); err != nil {
		t.Fatalf("auto migrate runtime nodes: %v", err)
	}

	module, err := BuildContainerRuntimeModule(root)
	if err == nil {
		t.Fatal("expected BuildContainerRuntimeModule() to fail when runtime agent dial fails")
	}
	if module != nil {
		t.Fatalf("expected module to be nil on runtime agent dial failure, got %+v", module)
	}
}

func TestBuildContainerRuntimeModuleFailsWhenLocalSandboxExecutorInitFails(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	cfg.RuntimeAgent.AllowLocalFallback = true

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}
	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}); err != nil {
		t.Fatalf("auto migrate runtime nodes: %v", err)
	}

	originalNewLocalRuntimeHostRunner := newLocalRuntimeHostRunner
	originalNewLocalSandboxExecutor := newLocalSandboxExecutor
	newLocalRuntimeHostRunner = func(*config.ContainerConfig) (runtimeports.RuntimeHostExecutor, error) {
		return newTestRuntimeEngine(zap.NewNop()), nil
	}
	newLocalSandboxExecutor = func(config.CheckerSandboxConfig) (runtimeports.SandboxExecutor, error) {
		return nil, errors.New("sandbox init failed")
	}
	t.Cleanup(func() {
		newLocalRuntimeHostRunner = originalNewLocalRuntimeHostRunner
		newLocalSandboxExecutor = originalNewLocalSandboxExecutor
	})

	module, err := BuildContainerRuntimeModule(root)
	if err == nil {
		t.Fatal("expected BuildContainerRuntimeModule() to fail when local checker runner init fails")
	}
	if module != nil {
		t.Fatalf("expected module to be nil on sandbox executor init failure, got %+v", module)
	}
	if err.Error() != "sandbox init failed" {
		t.Fatalf("error = %v, want sandbox init failed", err)
	}
}

func TestBuildContainerRuntimeModuleRejectsLocalRuntimeWithoutExplicitFallback(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.App.Env = "dev"

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	localRunnerCalled := false
	localSandboxCalled := false
	originalNewLocalRuntimeHostRunner := newLocalRuntimeHostRunner
	originalNewLocalSandboxExecutor := newLocalSandboxExecutor
	newLocalRuntimeHostRunner = func(*config.ContainerConfig) (runtimeports.RuntimeHostExecutor, error) {
		localRunnerCalled = true
		return &stubRuntimeNodeHostExecutor{}, nil
	}
	newLocalSandboxExecutor = func(config.CheckerSandboxConfig) (runtimeports.SandboxExecutor, error) {
		localSandboxCalled = true
		return stubSandboxExecutor{}, nil
	}
	t.Cleanup(func() {
		newLocalRuntimeHostRunner = originalNewLocalRuntimeHostRunner
		newLocalSandboxExecutor = originalNewLocalSandboxExecutor
	})

	module, err := BuildContainerRuntimeModule(root)
	if err == nil {
		t.Fatal("expected BuildContainerRuntimeModule() to reject local runtime without explicit fallback")
	}
	if module != nil {
		t.Fatalf("expected module to be nil on local runtime rejection, got %+v", module)
	}
	if !strings.Contains(err.Error(), "runtime_agent.allow_local_fallback") {
		t.Fatalf("error = %v, want runtime_agent.allow_local_fallback guidance", err)
	}
	if localRunnerCalled {
		t.Fatal("expected local runtime host runner not to be built")
	}
	if localSandboxCalled {
		t.Fatal("expected local sandbox executor not to be built")
	}
}

func TestBuildContainerRuntimeModuleAllowsExplicitLocalRuntimeFallback(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	cfg.RuntimeAgent.AllowLocalFallback = true

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	localRunnerCalled := false
	localSandboxCalled := false
	originalNewLocalRuntimeHostRunner := newLocalRuntimeHostRunner
	originalNewLocalSandboxExecutor := newLocalSandboxExecutor
	newLocalRuntimeHostRunner = func(*config.ContainerConfig) (runtimeports.RuntimeHostExecutor, error) {
		localRunnerCalled = true
		return &stubRuntimeNodeHostExecutor{}, nil
	}
	newLocalSandboxExecutor = func(config.CheckerSandboxConfig) (runtimeports.SandboxExecutor, error) {
		localSandboxCalled = true
		return stubSandboxExecutor{}, nil
	}
	t.Cleanup(func() {
		newLocalRuntimeHostRunner = originalNewLocalRuntimeHostRunner
		newLocalSandboxExecutor = originalNewLocalSandboxExecutor
	})

	module, err := BuildContainerRuntimeModule(root)
	if err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}
	if module == nil {
		t.Fatal("expected container runtime module")
	}
	if !localRunnerCalled {
		t.Fatal("expected explicit fallback to build local runtime host runner")
	}
	if !localSandboxCalled {
		t.Fatal("expected explicit fallback to build local sandbox executor")
	}
}

func TestRuntimeNodeClientAllowsTestRuntimeEngineWithoutFallback(t *testing.T) {
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	localRunnerCalled := false
	originalNewLocalRuntimeHostRunner := newLocalRuntimeHostRunner
	newLocalRuntimeHostRunner = func(*config.ContainerConfig) (runtimeports.RuntimeHostExecutor, error) {
		localRunnerCalled = true
		return &stubRuntimeNodeHostExecutor{}, nil
	}
	t.Cleanup(func() {
		newLocalRuntimeHostRunner = originalNewLocalRuntimeHostRunner
	})

	client, err := buildRuntimeNodeClientFromNode(context.Background(), cfg, zap.NewNop(), nil, nil)
	if err != nil {
		t.Fatalf("buildRuntimeNodeClientFromNode() error = %v", err)
	}
	if client == nil {
		t.Fatal("expected test runtime client")
	}
	if localRunnerCalled {
		t.Fatal("expected test runtime client not to build local Docker runner")
	}
}

func TestBuildContainerRuntimeModuleProvidesDefaultRuntimeNodeSelector(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	module, err := BuildContainerRuntimeModule(root)
	if err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}
	if module == nil || module.RuntimeNodeSelector == nil {
		t.Fatalf("expected runtime node selector, got %+v", module)
	}
	if module.OpsRuntimeQuery == nil {
		t.Fatalf("expected ops runtime query, got %+v", module)
	}

	if err := db.Create(&instanceentity.Instance{
		ID:          101,
		UserID:      9,
		ChallengeID: 21,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("seed running instance: %v", err)
	}
	activeContainers, err := module.OpsRuntimeQuery.CountRunning(context.Background())
	if err != nil {
		t.Fatalf("OpsRuntimeQuery.CountRunning() error = %v", err)
	}
	if activeContainers != 1 {
		t.Fatalf("OpsRuntimeQuery.CountRunning() = %d, want 1", activeContainers)
	}

	binding, err := module.RuntimeNodeSelector.SelectDefaultNode(context.Background())
	if err != nil {
		t.Fatalf("SelectDefaultNode() error = %v", err)
	}
	if binding == nil || binding.NodeID <= 0 {
		t.Fatalf("expected persisted default runtime node binding, got %+v", binding)
	}

	var stored containerruntimeentity.RuntimeNode
	if err := db.First(&stored, binding.NodeID).Error; err != nil {
		t.Fatalf("load runtime node: %v", err)
	}
	if !stored.Schedulable {
		t.Fatalf("expected default runtime node to be schedulable, got %+v", stored)
	}
}

func TestBuildContainerRuntimeModuleRegistersRuntimeNodeHealthJobWhenEnabled(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container.RuntimeNodeHealth.Enabled = true
	cfg.Container.RuntimeNodeHealth.PollInterval = time.Second
	cfg.Container.RuntimeNodeHealth.ProbeTimeout = 100 * time.Millisecond
	cfg.Container.RuntimeNodeHealth.StaleAfter = time.Minute
	cfg.Container.RuntimeNodeHealth.FailureThreshold = 1

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	module, err := BuildContainerRuntimeModule(root)
	if err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}
	if module == nil {
		t.Fatal("expected container runtime module")
	}

	if !backgroundJobRegistered(root, "runtime_node_health") {
		t.Fatalf("expected runtime_node_health background job, got %v", backgroundJobNames(root))
	}
}

func TestBuildContainerRuntimeModuleSkipsRuntimeNodeHealthJobWhenDisabled(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container.RuntimeNodeHealth.Enabled = false
	cfg.Container.RuntimeNodeHealth.PollInterval = time.Second
	cfg.Container.RuntimeNodeHealth.ProbeTimeout = 100 * time.Millisecond
	cfg.Container.RuntimeNodeHealth.StaleAfter = time.Minute
	cfg.Container.RuntimeNodeHealth.FailureThreshold = 1

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	if _, err := BuildContainerRuntimeModule(root); err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}

	if backgroundJobRegistered(root, "runtime_node_health") {
		t.Fatalf("expected runtime_node_health background job to be disabled, got %v", backgroundJobNames(root))
	}
}

func TestBuildDefaultRuntimeNodeSelectorRequiresFormalMigrationOutsideTestEnv(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	cfg.App.Env = "dev"

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	selector, repo, node, err := buildDefaultRuntimeNodeSelector(root, defaultRuntimeNodeName(cfg))
	if err == nil {
		t.Fatalf("expected missing runtime_nodes migration error, got selector=%+v repo=%+v node=%+v", selector, repo, node)
	}
	lowerErr := strings.ToLower(err.Error())
	if !strings.Contains(lowerErr, "no such table") && !strings.Contains(lowerErr, "does not exist") {
		t.Fatalf("expected missing table error, got %v", err)
	}
	if db.Migrator().HasTable(&containerruntimeentity.RuntimeNode{}) {
		t.Fatal("expected runtime node table to stay owned by formal SQL migrations")
	}
}

func TestBuildContainerRuntimeModuleSelectsConfiguredDefaultRuntimeNode(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:    true,
		Endpoint:   "runtime-agent.internal:7443",
		ServerName: "runtime-agent.internal",
	}

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	legacyNode := containerruntimeentity.RuntimeNode{
		Name:             "local-default",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC().Add(-time.Hour),
		UpdatedAt:        time.Now().UTC().Add(-time.Hour),
	}
	if err := db.Create(&legacyNode).Error; err != nil {
		t.Fatalf("create legacy runtime node: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	module, err := BuildContainerRuntimeModule(root)
	if err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}
	if module == nil || module.RuntimeNodeSelector == nil {
		t.Fatalf("expected runtime node selector, got %+v", module)
	}

	binding, err := module.RuntimeNodeSelector.SelectDefaultNode(context.Background())
	if err != nil {
		t.Fatalf("SelectDefaultNode() error = %v", err)
	}
	if binding == nil {
		t.Fatal("expected runtime node binding")
	}
	if binding.NodeName != "agent-default" {
		t.Fatalf("expected configured default node agent-default, got %+v", binding)
	}
	if binding.NodeID == legacyNode.ID {
		t.Fatalf("expected selector to avoid legacy local-default node, got %+v", binding)
	}
}

func TestBuildContainerRuntimeModuleSelectsConfiguredRuntimeAgentNodeName(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:    true,
		NodeName:   "runtime-node-a",
		Endpoint:   "runtime-agent.internal:7443",
		ServerName: "runtime-agent.internal",
	}

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	module, err := BuildContainerRuntimeModule(root)
	if err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}
	if module == nil || module.RuntimeNodeSelector == nil {
		t.Fatalf("expected runtime node selector, got %+v", module)
	}

	binding, err := module.RuntimeNodeSelector.SelectDefaultNode(context.Background())
	if err != nil {
		t.Fatalf("SelectDefaultNode() error = %v", err)
	}
	if binding == nil {
		t.Fatal("expected runtime node binding")
	}
	if binding.NodeName != "runtime-node-a" {
		t.Fatalf("expected configured runtime agent node runtime-node-a, got %+v", binding)
	}
}

func TestBuildContainerRuntimeModuleDefaultSelectorSkipsOfflineRuntimeNode(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container.RuntimeNodeHealth.Enabled = true
	cfg.Container.RuntimeNodeHealth.StaleAfter = time.Minute

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime module tables: %v", err)
	}

	now := time.Now().UTC()
	offlineDefault := containerruntimeentity.RuntimeNode{
		Name:             "local-default",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthOffline,
		CapacitySnapshot: "{}",
		LastSeenAt:       &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(&offlineDefault).Error; err != nil {
		t.Fatalf("create offline default runtime node: %v", err)
	}
	healthyNode := containerruntimeentity.RuntimeNode{
		Name:             "node-healthy",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		LastSeenAt:       &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(&healthyNode).Error; err != nil {
		t.Fatalf("create healthy runtime node: %v", err)
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	module, err := BuildContainerRuntimeModule(root)
	if err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}
	binding, err := module.RuntimeNodeSelector.SelectDefaultNode(context.Background())
	if err != nil {
		t.Fatalf("SelectDefaultNode() error = %v", err)
	}
	if binding == nil || binding.NodeID != healthyNode.ID {
		t.Fatalf("expected default selector to skip offline node and choose healthy node, got %+v", binding)
	}
}

func TestBuildContainerRuntimeModuleMigratesLegacyInstanceACLRulesToHandle(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime migration dependencies: %v", err)
	}

	node := containerruntimeentity.RuntimeNode{
		Name:             "local-default",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create runtime node: %v", err)
	}

	legacyDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		ACLRules: []runtimecontracts.InstanceRuntimeACLRule{
			{Comment: "ctf:acl:test", SourceIP: "172.30.0.2", TargetIP: "172.30.0.3", Action: "allow", Protocol: "tcp", Ports: []int{3306}},
		},
	})
	if err != nil {
		t.Fatalf("encode legacy runtime details: %v", err)
	}

	nodeID := node.ID
	instance := instanceentity.Instance{
		ID:             1001,
		UserID:         2001,
		ChallengeID:    3001,
		RuntimeNodeID:  &nodeID,
		RuntimeDetails: legacyDetails,
		ShareScope:     instanceentity.ShareScopePerUser,
		Status:         instanceentity.InstanceStatusRunning,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	executor := &stubRuntimeNodeHostExecutor{}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		node.ID: newStubNodeRuntimeClient(executor, nil),
	})

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	if _, err := BuildContainerRuntimeModule(root); err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}

	if len(executor.appliedACLCalls) != 1 {
		t.Fatalf("expected 1 apply acl migration call, got %+v", executor.appliedACLCalls)
	}
	if len(executor.removedACLRulesCalls) != 1 {
		t.Fatalf("expected 1 legacy acl removal call, got %+v", executor.removedACLRulesCalls)
	}

	var stored instanceentity.Instance
	if err := db.First(&stored, instance.ID).Error; err != nil {
		t.Fatalf("reload instance: %v", err)
	}
	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(stored.RuntimeDetails)
	if err != nil {
		t.Fatalf("decode migrated runtime details: %v", err)
	}
	if details.ACL == nil || details.ACL.Chain != "CTF-INS-1001" {
		t.Fatalf("expected acl handle CTF-INS-1001, got %+v", details.ACL)
	}
	if len(details.ACLRules) != 1 || details.ACLRules[0].Comment != "ctf:acl:test" {
		t.Fatalf("expected acl rules snapshot preserved, got %+v", details.ACLRules)
	}
}

func TestBuildContainerRuntimeModuleIgnoresMissingLegacyACLRuleDuringMigration(t *testing.T) {
	cfg, db, cache := newRootTestDependencies(t)

	if err := db.AutoMigrate(&containerruntimeentity.RuntimeNode{}, &instanceentity.Instance{}); err != nil {
		t.Fatalf("auto migrate runtime migration dependencies: %v", err)
	}

	node := containerruntimeentity.RuntimeNode{
		Name:             "local-default",
		Endpoint:         "local://docker",
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     containerruntimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create runtime node: %v", err)
	}

	legacyDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		ACLRules: []runtimecontracts.InstanceRuntimeACLRule{
			{Comment: "ctf:acl:test", SourceIP: "172.30.0.2", TargetIP: "172.30.0.3", Action: "allow", Protocol: "tcp", Ports: []int{3306}},
		},
	})
	if err != nil {
		t.Fatalf("encode legacy runtime details: %v", err)
	}

	nodeID := node.ID
	instance := instanceentity.Instance{
		ID:             1002,
		UserID:         2002,
		ChallengeID:    3002,
		RuntimeNodeID:  &nodeID,
		RuntimeDetails: legacyDetails,
		ShareScope:     instanceentity.ShareScopePerUser,
		Status:         instanceentity.InstanceStatusRunning,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	executor := &stubRuntimeNodeHostExecutor{
		removeACLRulesErr: errors.New("iptables -D DOCKER-USER failed: does a matching rule exist in that chain?"),
	}
	overrideRuntimeNodeClientBuilder(t, map[int64]runtimeNodeClient{
		node.ID: newStubNodeRuntimeClient(executor, nil),
	})

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	if _, err := BuildContainerRuntimeModule(root); err != nil {
		t.Fatalf("BuildContainerRuntimeModule() error = %v", err)
	}

	var stored instanceentity.Instance
	if err := db.First(&stored, instance.ID).Error; err != nil {
		t.Fatalf("reload instance: %v", err)
	}
	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(stored.RuntimeDetails)
	if err != nil {
		t.Fatalf("decode migrated runtime details: %v", err)
	}
	if details.ACL == nil || details.ACL.Chain != "CTF-INS-1002" {
		t.Fatalf("expected acl handle CTF-INS-1002, got %+v", details.ACL)
	}
}

type stubRuntimeHTTPProxyTickets struct {
	ticket    string
	expiresAt time.Time
}

func (s stubRuntimeHTTPProxyTickets) IssueTicket(context.Context, authctx.CurrentUser, int64) (string, time.Time, error) {
	return s.ticket, s.expiresAt, nil
}

func (s stubRuntimeHTTPProxyTickets) IssueAWDTargetTicket(context.Context, authctx.CurrentUser, int64, int64, int64) (string, time.Time, error) {
	return s.ticket, s.expiresAt, nil
}

func (s stubRuntimeHTTPProxyTickets) IssueAWDDefenseSSHTicket(context.Context, authctx.CurrentUser, int64, int64) (string, time.Time, error) {
	return s.ticket, s.expiresAt, nil
}

func (s stubRuntimeHTTPProxyTickets) ResolveTicket(context.Context, string) (*instanceports.ProxyTicketClaims, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTickets) ResolveAWDTargetAccessURL(context.Context, *instanceports.ProxyTicketClaims, int64, int64, int64) (string, error) {
	return "", nil
}

func (s stubRuntimeHTTPProxyTickets) MaxAge() int {
	return 900
}

type stubSandboxExecutor struct{}

func (stubSandboxExecutor) RunSandboxExec(context.Context, runtimeports.SandboxExecJob) (runtimeports.SandboxExecResult, error) {
	return runtimeports.SandboxExecResult{Status: runtimeports.SandboxExecStatusOK}, nil
}

func writeRemoteAgentClientTLSFiles(t *testing.T) (string, string, string) {
	t.Helper()

	dir := t.TempDir()
	certPEM, keyPEM := newSelfSignedClientCertificatePEM(t, "runtime-agent.local")
	caFile := filepath.Join(dir, "ca.pem")
	certFile := filepath.Join(dir, "client.pem")
	keyFile := filepath.Join(dir, "client-key.pem")

	for _, file := range []struct {
		path string
		data []byte
	}{
		{path: caFile, data: certPEM},
		{path: certFile, data: certPEM},
		{path: keyFile, data: keyPEM},
	} {
		if err := os.WriteFile(file.path, file.data, 0o600); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", file.path, err)
		}
	}

	return caFile, certFile, keyFile
}

func newSelfSignedClientCertificatePEM(t *testing.T, commonName string) ([]byte, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: bigIntFromTime(time.Now()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IsCA:        true,
		DNSNames:    []string{commonName},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
}

func bigIntFromTime(ts time.Time) *big.Int {
	return big.NewInt(ts.UnixNano())
}

func backgroundJobRegistered(root *Root, name string) bool {
	for _, job := range root.BackgroundJobs() {
		if job.Name() == name {
			return true
		}
	}
	return false
}

func backgroundJobNames(root *Root) []string {
	jobs := root.BackgroundJobs()
	names := make([]string, 0, len(jobs))
	for _, job := range jobs {
		names = append(names, job.Name())
	}
	return names
}
