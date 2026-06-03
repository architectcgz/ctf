package composition

import (
	"context"
	"strings"
	"sync"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type runtimeNodeClient interface {
	CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error)
	CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error
	RemoveContainer(ctx context.Context, containerID string) error
	InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error)
	ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error)
	StartContainer(ctx context.Context, containerID string) error
	RunChecker(ctx context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	Close(ctx context.Context) error
}

type nodeRuntimeClient struct {
	executor      runtimeports.RuntimeHostExecutor
	checkerRunner contestports.CheckerRunner
	provisioner   *runtimecmd.ProvisioningService
	cleaner       *runtimecmd.RuntimeCleanupService
	closer        runtimeLifecycleCloser
}

type runtimeNodeExecutionRouter struct {
	cfg             *config.Config
	logger          *zap.Logger
	runtimeRepo     *runtimeinfra.Repository
	nodeRepo        *runtimeinfra.RuntimeNodeRepository
	defaultNodeName string

	mu               sync.Mutex
	clients          map[int64]runtimeNodeClient
	containerNodeIDs map[string]int64
}

var buildRuntimeNodeClient = buildRuntimeNodeClientFromNode

func newNodeRuntimeClient(
	cfg *config.Config,
	logger *zap.Logger,
	runtimeRepo *runtimeinfra.Repository,
	executor runtimeports.RuntimeHostExecutor,
	checkerRunner contestports.CheckerRunner,
	closer runtimeLifecycleCloser,
) *nodeRuntimeClient {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &nodeRuntimeClient{
		executor:      executor,
		checkerRunner: checkerRunner,
		provisioner:   runtimecmd.NewProvisioningService(runtimeRepo, executor, &cfg.Container, logger.Named("runtime_provisioning_service")),
		cleaner:       runtimecmd.NewRuntimeCleanupService(executor, runtimeRepo, logger.Named("runtime_cleanup_service")),
		closer:        closer,
	}
}

func buildRuntimeNodeClientFromNode(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	runtimeRepo *runtimeinfra.Repository,
	node *runtimeentity.RuntimeNode,
) (runtimeNodeClient, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg.App.Env == "test" {
		executor := newTestRuntimeEngine(logger.Named("runtime_test_engine"))
		return newNodeRuntimeClient(cfg, logger, runtimeRepo, executor, nil, nil), nil
	}
	if node == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	if usesLocalRuntimeNode(node) {
		executor, err := newLocalRuntimeHostRunner(&cfg.Container)
		if err != nil {
			return nil, err
		}
		runner, err := newLocalCheckerRunner(cfg.Contest.AWD.CheckerSandbox)
		if err != nil {
			return nil, err
		}
		return newNodeRuntimeClient(cfg, logger, runtimeRepo, executor, runner, nil), nil
	}

	bridge, err := dialRuntimeAgent(ctx, runtimeAgentConfigForNode(cfg.RuntimeAgent, node))
	if err != nil {
		return nil, err
	}
	return newNodeRuntimeClient(cfg, logger, runtimeRepo, bridge, bridge, bridge), nil
}

func newRuntimeNodeExecutionRouter(
	cfg *config.Config,
	logger *zap.Logger,
	runtimeRepo *runtimeinfra.Repository,
	nodeRepo *runtimeinfra.RuntimeNodeRepository,
	defaultNodeName string,
) *runtimeNodeExecutionRouter {
	if runtimeRepo == nil || nodeRepo == nil {
		return nil
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &runtimeNodeExecutionRouter{
		cfg:              cfg,
		logger:           logger,
		runtimeRepo:      runtimeRepo,
		nodeRepo:         nodeRepo,
		defaultNodeName:  strings.TrimSpace(defaultNodeName),
		clients:          make(map[int64]runtimeNodeClient),
		containerNodeIDs: make(map[string]int64),
	}
}

func (r *runtimeNodeExecutionRouter) rememberClient(nodeID int64, client runtimeNodeClient) {
	if r == nil || nodeID <= 0 || client == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[nodeID] = client
}

func (r *runtimeNodeExecutionRouter) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	clients := make([]runtimeNodeClient, 0, len(r.clients))
	for _, client := range r.clients {
		if client == nil {
			continue
		}
		clients = append(clients, client)
	}
	r.mu.Unlock()

	var closeErr error
	for _, client := range clients {
		if err := client.Close(ctx); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func (c *nodeRuntimeClient) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if c == nil || c.provisioner == nil || req == nil {
		return nil, nil
	}
	result, err := c.provisioner.CreateTopology(ctx, toRuntimeTopologyCreateRequestFromPractice(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResultForPractice(result), nil
}

func (r *runtimeNodeExecutionRouter) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if r == nil || req == nil {
		return nil, nil
	}
	client, _, err := r.clientForNodeID(ctx, req.NodeID)
	if err != nil {
		return nil, err
	}
	return client.CreateTopology(ctx, req)
}

func (c *nodeRuntimeClient) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
	if c == nil || c.provisioner == nil {
		return "", "", 0, 0, nil
	}
	return c.provisioner.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func (r *runtimeNodeExecutionRouter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, nodeID int64) (string, string, int, int, error) {
	if r == nil {
		return "", "", 0, 0, nil
	}
	client, _, err := r.clientForNodeID(ctx, nodeID)
	if err != nil {
		return "", "", 0, 0, err
	}
	return client.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func (c *nodeRuntimeClient) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if c == nil || c.cleaner == nil || instance == nil {
		return nil
	}
	return c.cleaner.CleanupRuntime(ctx, instance)
}

func (r *runtimeNodeExecutionRouter) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if r == nil || instance == nil {
		return nil
	}
	client, _, err := r.clientForCleanupRuntime(ctx, instance)
	if err != nil {
		return err
	}
	return client.CleanupRuntime(ctx, instance)
}

func (c *nodeRuntimeClient) RemoveContainer(ctx context.Context, containerID string) error {
	if c == nil || c.cleaner == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	return c.cleaner.RemoveContainer(ctx, containerID)
}

func (r *runtimeNodeExecutionRouter) RemoveContainer(ctx context.Context, containerID string) error {
	if r == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	client, _, err := r.clientForContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	return client.RemoveContainer(ctx, containerID)
}

func (c *nodeRuntimeClient) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if c == nil || c.executor == nil || strings.TrimSpace(containerID) == "" {
		return nil, nil
	}
	return c.executor.InspectManagedContainer(ctx, containerID)
}

func (r *runtimeNodeExecutionRouter) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if r == nil || strings.TrimSpace(containerID) == "" {
		return nil, nil
	}
	client, _, err := r.clientForContainerID(ctx, containerID)
	if err != nil {
		return nil, err
	}
	return client.InspectManagedContainer(ctx, containerID)
}

func (c *nodeRuntimeClient) ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error) {
	if c == nil || c.executor == nil {
		return nil, nil
	}
	return c.executor.ListManagedContainers(ctx)
}

func (r *runtimeNodeExecutionRouter) ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error) {
	if r == nil {
		return nil, nil
	}
	nodes, err := r.nodeRepo.ListSchedulableNodes(ctx)
	if err != nil {
		return nil, err
	}

	containers := make([]runtimeports.ManagedContainer, 0)
	seen := make(map[string]struct{})
	for i := range nodes {
		client, nodeID, err := r.clientForConcreteNode(ctx, &nodes[i])
		if err != nil {
			return nil, err
		}
		items, err := client.ListManagedContainers(ctx)
		if err != nil {
			return nil, err
		}
		r.recordContainerNodeIDs(nodeID, items)
		for _, item := range items {
			if strings.TrimSpace(item.ID) == "" {
				continue
			}
			if _, exists := seen[item.ID]; exists {
				continue
			}
			seen[item.ID] = struct{}{}
			containers = append(containers, item)
		}
	}
	return containers, nil
}

func (c *nodeRuntimeClient) StartContainer(ctx context.Context, containerID string) error {
	if c == nil || c.executor == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	return c.executor.StartContainer(ctx, containerID)
}

func (r *runtimeNodeExecutionRouter) StartContainer(ctx context.Context, containerID string) error {
	if r == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	client, _, err := r.clientForContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	return client.StartContainer(ctx, containerID)
}

func (c *nodeRuntimeClient) RunChecker(ctx context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	if c == nil || c.checkerRunner == nil {
		return contestports.CheckerRunResult{}, nil
	}
	return c.checkerRunner.RunChecker(ctx, job)
}

func (r *runtimeNodeExecutionRouter) RunChecker(ctx context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	if r == nil {
		return contestports.CheckerRunResult{}, nil
	}
	client, _, err := r.clientForNodeID(ctx, job.Metadata.NodeID)
	if err != nil {
		return contestports.CheckerRunResult{}, err
	}
	return client.RunChecker(ctx, job)
}

func (c *nodeRuntimeClient) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	if c == nil || c.executor == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	return c.executor.WriteFileToContainer(ctx, containerID, filePath, content)
}

func (r *runtimeNodeExecutionRouter) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	if r == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	client, _, err := r.clientForContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	return client.WriteFileToContainer(ctx, containerID, filePath, content)
}

func (c *nodeRuntimeClient) Close(ctx context.Context) error {
	if c == nil || c.closer == nil {
		return nil
	}
	return c.closer.Close(ctx)
}

func (r *runtimeNodeExecutionRouter) clientForInstance(ctx context.Context, instance *instancecontracts.Instance) (runtimeNodeClient, int64, error) {
	if instance == nil {
		return nil, 0, nil
	}
	return r.clientForNodeID(ctx, runtimeNodeIDValue(instance.NodeID))
}

func (r *runtimeNodeExecutionRouter) clientForCleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) (runtimeNodeClient, int64, error) {
	if instance == nil {
		return nil, 0, nil
	}
	if nodeID := runtimeNodeIDValue(instance.NodeID); nodeID > 0 {
		return r.clientForNodeID(ctx, nodeID)
	}
	for _, containerID := range cleanupRuntimeContainerIDs(instance) {
		nodeID, err := r.resolveNodeIDForContainer(ctx, containerID)
		if err != nil {
			return nil, 0, err
		}
		if nodeID <= 0 {
			continue
		}
		return r.clientForNodeID(ctx, nodeID)
	}
	return r.clientForInstance(ctx, instance)
}

func (r *runtimeNodeExecutionRouter) clientForContainerID(ctx context.Context, containerID string) (runtimeNodeClient, int64, error) {
	nodeID, err := r.resolveNodeIDForContainer(ctx, containerID)
	if err != nil {
		return nil, 0, err
	}
	return r.clientForNodeID(ctx, nodeID)
}

func (r *runtimeNodeExecutionRouter) clientForNodeID(ctx context.Context, nodeID int64) (runtimeNodeClient, int64, error) {
	node, err := r.resolveNodeForExecution(ctx, nodeID)
	if err != nil {
		return nil, 0, err
	}
	return r.clientForConcreteNode(ctx, node)
}

func (r *runtimeNodeExecutionRouter) clientForConcreteNode(ctx context.Context, node *runtimeentity.RuntimeNode) (runtimeNodeClient, int64, error) {
	if r == nil || node == nil || node.ID <= 0 {
		return nil, 0, runtimeports.ErrRuntimeNodeUnavailable
	}

	r.mu.Lock()
	if client, ok := r.clients[node.ID]; ok && client != nil {
		r.mu.Unlock()
		return client, node.ID, nil
	}
	r.mu.Unlock()

	client, err := buildRuntimeNodeClient(ctx, r.cfg, r.logger, r.runtimeRepo, node)
	if err != nil {
		return nil, 0, err
	}

	r.mu.Lock()
	if existing, ok := r.clients[node.ID]; ok && existing != nil {
		r.mu.Unlock()
		_ = client.Close(context.Background())
		return existing, node.ID, nil
	}
	r.clients[node.ID] = client
	r.mu.Unlock()
	return client, node.ID, nil
}

func (r *runtimeNodeExecutionRouter) resolveNodeForExecution(ctx context.Context, nodeID int64) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.nodeRepo == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	if nodeID > 0 {
		return r.nodeRepo.FindByID(ctx, nodeID)
	}
	if r.defaultNodeName != "" {
		return r.nodeRepo.FindSchedulableNodeByName(ctx, r.defaultNodeName)
	}
	return r.nodeRepo.FindFirstSchedulableNode(ctx)
}

func (r *runtimeNodeExecutionRouter) resolveNodeIDForContainer(ctx context.Context, containerID string) (int64, error) {
	trimmedID := strings.TrimSpace(containerID)
	if trimmedID == "" {
		return 0, nil
	}

	r.mu.Lock()
	cachedNodeID := r.containerNodeIDs[trimmedID]
	r.mu.Unlock()
	if cachedNodeID > 0 {
		return cachedNodeID, nil
	}

	if r.runtimeRepo != nil {
		nodeID, err := r.runtimeRepo.FindRuntimeNodeIDByContainerID(ctx, trimmedID)
		if err != nil {
			return 0, err
		}
		if nodeID != nil && *nodeID > 0 {
			r.mu.Lock()
			r.containerNodeIDs[trimmedID] = *nodeID
			r.mu.Unlock()
			return *nodeID, nil
		}
	}
	return 0, nil
}

func (r *runtimeNodeExecutionRouter) recordContainerNodeIDs(nodeID int64, containers []runtimeports.ManagedContainer) {
	if r == nil || nodeID <= 0 || len(containers) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range containers {
		containerID := strings.TrimSpace(item.ID)
		if containerID == "" {
			continue
		}
		r.containerNodeIDs[containerID] = nodeID
	}
}

func runtimeNodeIDValue(nodeID *int64) int64 {
	if nodeID == nil || *nodeID <= 0 {
		return 0
	}
	return *nodeID
}

func cleanupRuntimeContainerIDs(instance *instancecontracts.Instance) []string {
	if instance == nil {
		return nil
	}

	result := make([]string, 0, 4)
	seen := make(map[string]struct{}, 4)
	appendContainerID := func(containerID string) {
		containerID = strings.TrimSpace(containerID)
		if containerID == "" {
			return
		}
		if _, exists := seen[containerID]; exists {
			return
		}
		seen[containerID] = struct{}{}
		result = append(result, containerID)
	}

	appendContainerID(instance.ContainerID)

	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil {
		return result
	}
	for _, item := range details.Containers {
		appendContainerID(item.ContainerID)
	}
	return result
}

func usesLocalRuntimeNode(node *runtimeentity.RuntimeNode) bool {
	if node == nil {
		return true
	}
	return strings.TrimSpace(node.Endpoint) == "" || strings.TrimSpace(node.Endpoint) == "local://docker"
}

func runtimeAgentConfigForNode(base config.RuntimeAgentConfig, node *runtimeentity.RuntimeNode) config.RuntimeAgentConfig {
	cfg := base
	cfg.Enabled = true
	if node != nil {
		cfg.Endpoint = strings.TrimSpace(node.Endpoint)
		if tlsIdentity := strings.TrimSpace(node.TLSIdentity); tlsIdentity != "" {
			cfg.ServerName = tlsIdentity
		}
	}
	return cfg
}

var _ contestports.CheckerRunner = (*runtimeNodeExecutionRouter)(nil)
var _ contestports.AWDContainerFileWriter = (*runtimeNodeExecutionRouter)(nil)
var _ practiceManagedContainerInspector = (*runtimeNodeExecutionRouter)(nil)
var _ runtimeNodeClient = (*nodeRuntimeClient)(nil)
