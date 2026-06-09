package composition

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

const (
	runtimeNodeRouterE2EImage                  = "nginx:1.27-alpine"
	runtimeNodeRouterE2ECreateTimeout          = 2 * time.Minute
	runtimeNodeRouterE2ECleanupTimeout         = 30 * time.Second
	runtimeNodeRouterE2EContainerRemovalWait   = 30 * time.Second
	runtimeNodeRouterE2EContainerRemovalPoll   = 500 * time.Millisecond
	runtimeNodeRouterE2EDockerInspectTimeout   = 5 * time.Second
	runtimeNodeRouterE2EDefaultRuntimeNodeName = "agent-a-e2e"
)

type runtimeNodeRouterE2EEnv struct {
	agentAEndpoint   string
	agentAServerName string
	agentBEndpoint   string
	agentBServerName string
	caFile           string
	clientCertFile   string
	clientKeyFile    string
	dockerHostA      string
	dockerHostB      string
	image            string
}

func TestRuntimeNodeExecutionRouterE2ECleanupUsesRuntimeDetailsContainerNode(t *testing.T) {
	env, ok := loadRuntimeNodeRouterE2EEnv(t)
	if !ok {
		t.Skip("runtime node router e2e env not configured")
	}

	cfg, db, _ := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:     true,
		Endpoint:    env.agentAEndpoint,
		DialTimeout: 5 * time.Second,
		ServerName:  env.agentAServerName,
		CAFile:      env.caFile,
		CertFile:    env.clientCertFile,
		KeyFile:     env.clientKeyFile,
	}
	applyRuntimeNodeRouterE2EContainerConfig(&cfg.Container)

	if err := db.AutoMigrate(
		&runtimeentity.RuntimeNode{},
		&instanceentity.Instance{},
		&runtimeentity.PortAllocation{},
		&runtimeentity.NetworkAllocation{},
	); err != nil {
		t.Fatalf("auto migrate runtime router e2e tables: %v", err)
	}

	_, nodeB := seedRuntimeRouterE2ENodes(t, db, env)
	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		runtimeinfra.NewAllocationRepository(db),
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
		runtimeNodeRouterE2EDefaultRuntimeNodeName,
	)
	t.Cleanup(func() {
		if router != nil {
			_ = router.Close(context.Background())
		}
	})

	dockerA := newRuntimeNodeRouterE2EDockerClient(t, env.dockerHostA)
	dockerB := newRuntimeNodeRouterE2EDockerClient(t, env.dockerHostB)

	ctx, cancel := context.WithTimeout(context.Background(), runtimeNodeRouterE2ECreateTimeout)
	defer cancel()

	containerID, networkID, hostPort, servicePort, err := router.CreateContainer(ctx, env.image, map[string]string{"CTF_ROUTER_E2E": "runtime-details"}, 0, nodeB.ID)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	assertContainerPresent(t, dockerB, containerID)
	assertContainerAbsent(t, dockerA, containerID)

	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Networks: []runtimecontracts.InstanceRuntimeNetwork{
			{
				Key:       runtimecontracts.TopologyDefaultNetworkKey,
				Name:      "ctf-e2e-runtime-details",
				NetworkID: networkID,
			},
		},
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{
				NodeKey:         "default",
				ContainerID:     containerID,
				HostPort:        hostPort,
				ServicePort:     servicePort,
				ServiceProtocol: runtimecontracts.ChallengeTargetProtocolHTTP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
			},
		},
	})
	if err != nil {
		t.Fatalf("EncodeInstanceRuntimeDetails() error = %v", err)
	}

	nodeBID := nodeB.ID
	storedInstance := instanceentity.Instance{
		ID:             5001,
		UserID:         6001,
		ChallengeID:    7001,
		NodeID:         &nodeBID,
		RuntimeDetails: runtimeDetails,
		HostPort:       hostPort,
		Status:         instanceentity.InstanceStatusRunning,
		ShareScope:     instanceentity.ShareScopePerUser,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := db.Create(&storedInstance).Error; err != nil {
		t.Fatalf("create stored instance: %v", err)
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), runtimeNodeRouterE2ECleanupTimeout)
	defer cleanupCancel()
	if err := router.CleanupRuntime(cleanupCtx, runtimeCleanupTargetFromInstance(&instanceentity.Instance{RuntimeDetails: runtimeDetails})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	waitForContainerAbsent(t, dockerB, containerID)
	assertContainerAbsent(t, dockerA, containerID)
}

func TestRuntimeNodeExecutionRouterE2ECleanupUsesWorkspaceContainerNode(t *testing.T) {
	env, ok := loadRuntimeNodeRouterE2EEnv(t)
	if !ok {
		t.Skip("runtime node router e2e env not configured")
	}

	cfg, db, _ := newRootTestDependencies(t)
	cfg.App.Env = "dev"
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:     true,
		Endpoint:    env.agentAEndpoint,
		DialTimeout: 5 * time.Second,
		ServerName:  env.agentAServerName,
		CAFile:      env.caFile,
		CertFile:    env.clientCertFile,
		KeyFile:     env.clientKeyFile,
	}
	applyRuntimeNodeRouterE2EContainerConfig(&cfg.Container)

	if err := db.AutoMigrate(
		&runtimeentity.RuntimeNode{},
		&instanceentity.Instance{},
		&runtimeentity.AWDDefenseWorkspace{},
		&runtimeentity.PortAllocation{},
		&runtimeentity.NetworkAllocation{},
	); err != nil {
		t.Fatalf("auto migrate workspace e2e tables: %v", err)
	}

	_, nodeB := seedRuntimeRouterE2ENodes(t, db, env)
	router := newRuntimeNodeExecutionRouter(
		cfg,
		zap.NewNop(),
		runtimeinfra.NewAllocationRepository(db),
		runtimeinfra.NewRepository(db),
		runtimeinfra.NewRuntimeNodeRepository(db),
		runtimeNodeRouterE2EDefaultRuntimeNodeName,
	)
	t.Cleanup(func() {
		if router != nil {
			_ = router.Close(context.Background())
		}
	})

	dockerA := newRuntimeNodeRouterE2EDockerClient(t, env.dockerHostA)
	dockerB := newRuntimeNodeRouterE2EDockerClient(t, env.dockerHostB)

	ctx, cancel := context.WithTimeout(context.Background(), runtimeNodeRouterE2ECreateTimeout)
	defer cancel()

	containerID, _, _, _, err := router.CreateContainer(ctx, env.image, map[string]string{"CTF_ROUTER_E2E": "workspace"}, 0, nodeB.ID)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	assertContainerPresent(t, dockerB, containerID)
	assertContainerAbsent(t, dockerA, containerID)

	nodeBID := nodeB.ID
	storedInstance := instanceentity.Instance{
		ID:          5101,
		UserID:      6101,
		ContestID:   int64PtrForRouterTest(8101),
		TeamID:      int64PtrForRouterTest(8201),
		ServiceID:   int64PtrForRouterTest(8301),
		ChallengeID: 7101,
		NodeID:      &nodeBID,
		ContainerID: "primary-node-b",
		Status:      instanceentity.InstanceStatusRunning,
		ShareScope:  instanceentity.ShareScopePerUser,
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := db.Create(&storedInstance).Error; err != nil {
		t.Fatalf("create workspace owner instance: %v", err)
	}
	workspace := runtimeentity.AWDDefenseWorkspace{
		ContestID:         8101,
		TeamID:            8201,
		ServiceID:         8301,
		InstanceID:        storedInstance.ID,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       containerID,
		SeedSignature:     "runtime-node-router-e2e",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("create workspace row: %v", err)
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), runtimeNodeRouterE2ECleanupTimeout)
	defer cleanupCancel()
	if err := router.CleanupRuntime(cleanupCtx, runtimeCleanupTargetFromInstance(&instanceentity.Instance{ContainerID: containerID})); err != nil {
		t.Fatalf("CleanupRuntime() error = %v", err)
	}

	waitForContainerAbsent(t, dockerB, containerID)
	assertContainerAbsent(t, dockerA, containerID)
}

func loadRuntimeNodeRouterE2EEnv(t *testing.T) (runtimeNodeRouterE2EEnv, bool) {
	t.Helper()

	env := runtimeNodeRouterE2EEnv{
		agentAEndpoint:   strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_AGENT_A_ENDPOINT")),
		agentAServerName: strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_AGENT_A_SERVER_NAME")),
		agentBEndpoint:   strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_AGENT_B_ENDPOINT")),
		agentBServerName: strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_AGENT_B_SERVER_NAME")),
		caFile:           strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_CA_FILE")),
		clientCertFile:   strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_CLIENT_CERT_FILE")),
		clientKeyFile:    strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_CLIENT_KEY_FILE")),
		dockerHostA:      strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_DOCKER_HOST_A")),
		dockerHostB:      strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_DOCKER_HOST_B")),
		image:            strings.TrimSpace(os.Getenv("CTF_RUNTIME_ROUTER_E2E_IMAGE")),
	}
	if env.image == "" {
		env.image = runtimeNodeRouterE2EImage
	}
	required := []string{
		env.agentAEndpoint,
		env.agentAServerName,
		env.agentBEndpoint,
		env.agentBServerName,
		env.caFile,
		env.clientCertFile,
		env.clientKeyFile,
		env.dockerHostA,
		env.dockerHostB,
	}
	for _, item := range required {
		if item == "" {
			return runtimeNodeRouterE2EEnv{}, false
		}
	}
	return env, true
}

func seedRuntimeRouterE2ENodes(t *testing.T, db *gorm.DB, env runtimeNodeRouterE2EEnv) (*runtimeentity.RuntimeNode, *runtimeentity.RuntimeNode) {
	t.Helper()

	nodeA := &runtimeentity.RuntimeNode{
		Name:             runtimeNodeRouterE2EDefaultRuntimeNodeName,
		Endpoint:         env.agentAEndpoint,
		TLSIdentity:      env.agentAServerName,
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     runtimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(nodeA).Error; err != nil {
		t.Fatalf("create node A: %v", err)
	}
	nodeB := &runtimeentity.RuntimeNode{
		Name:             "agent-b-e2e",
		Endpoint:         env.agentBEndpoint,
		TLSIdentity:      env.agentBServerName,
		Schedulable:      true,
		Labels:           "{}",
		HealthStatus:     runtimeentity.RuntimeNodeHealthReady,
		CapacitySnapshot: "{}",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(nodeB).Error; err != nil {
		t.Fatalf("create node B: %v", err)
	}
	return nodeA, nodeB
}

func applyRuntimeNodeRouterE2EContainerConfig(cfg *config.ContainerConfig) {
	if cfg == nil {
		return
	}
	cfg.DefaultCPUQuota = 0.5
	cfg.DefaultMemory = 256 * 1024 * 1024
	cfg.DefaultPidsLimit = 256
	cfg.PortRangeStart = 32000
	cfg.PortRangeEnd = 32100
	cfg.DefaultExposedPort = 80
	cfg.PublicHost = "127.0.0.1"
}

func newRuntimeNodeRouterE2EDockerClient(t *testing.T, host string) *client.Client {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("new docker client for %s: %v", host, err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	return cli
}

func assertContainerPresent(t *testing.T, cli *client.Client, containerID string) {
	t.Helper()

	if !containerPresent(t, cli, containerID) {
		t.Fatalf("expected container %s to exist on %s", containerID, cli.DaemonHost())
	}
}

func assertContainerAbsent(t *testing.T, cli *client.Client, containerID string) {
	t.Helper()

	if containerPresent(t, cli, containerID) {
		t.Fatalf("expected container %s to be absent on %s", containerID, cli.DaemonHost())
	}
}

func waitForContainerAbsent(t *testing.T, cli *client.Client, containerID string) {
	t.Helper()

	deadline := time.Now().Add(runtimeNodeRouterE2EContainerRemovalWait)
	for time.Now().Before(deadline) {
		if !containerPresent(t, cli, containerID) {
			return
		}
		time.Sleep(runtimeNodeRouterE2EContainerRemovalPoll)
	}
	t.Fatalf("container %s still exists on %s after cleanup wait", containerID, cli.DaemonHost())
}

func containerPresent(t *testing.T, cli *client.Client, containerID string) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), runtimeNodeRouterE2EDockerInspectTimeout)
	defer cancel()
	_, err := cli.ContainerInspect(ctx, containerID)
	if err == nil {
		return true
	}
	if errdefs.IsNotFound(err) {
		return false
	}
	t.Fatalf("inspect container %s on %s: %v", containerID, cli.DaemonHost(), err)
	return false
}
