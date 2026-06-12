package composition

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	challengeports "ctf-platform/internal/module/challenge/ports"
	runtimeapp "ctf-platform/internal/module/container_runtime/application"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	"ctf-platform/internal/module/container_runtime/infrastructure/agentclient"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	containerruntime "ctf-platform/internal/module/container_runtime/runtime"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	opsports "ctf-platform/internal/module/ops/ports"
)

type runtimeLifecycleCloser interface {
	Close(ctx context.Context) error
}

var (
	dialRuntimeAgent          = agentclient.DialContext
	newLocalRuntimeHostRunner = containerruntimeinfra.NewLocalHostExecutor
	newLocalSandboxExecutor   = func(cfg config.CheckerSandboxConfig) (runtimeports.SandboxExecutor, error) {
		return containerruntimeinfra.NewDockerSandboxExecutor(cfg)
	}
)

type ContainerRuntimeModule struct {
	ChallengeImageRuntime   challengeports.ImageRuntime
	ChallengeRuntimeProbe   challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery         opsports.RuntimeQuery
	OpsRuntimeStatsProvider opsports.RuntimeStatsProvider
	ContestContainerFiles   contestports.AWDContainerFileWriter
	ContestCheckerRunner    contestports.CheckerRunner
	RuntimeNodeSelector     runtimeports.RuntimeNodeSelector
	LifecycleCloser         runtimeLifecycleCloser

	nodeRouter        *runtimeNodeExecutionRouter
	runtime           *containerruntime.Module
	runtimeNodeHealth *runtimeapp.NodeHealthService
}

func BuildContainerRuntimeModule(root *Root) (*ContainerRuntimeModule, error) {
	cfg := runtimeConfigOrDefault(root.Config())
	log := root.Logger()
	indexRepo := newCompositeRuntimeNodeContainerIndex(
		newInstanceRuntimeInventoryProvider(instanceinfra.NewContainerInventoryRepository(root.DB())),
		contestinfra.NewAWDContainerInventoryRepository(root.DB()),
	)
	aclMigrationRepo := instanceinfra.NewACLMigrationStateRepository(root.DB())
	allocationRepo := containerruntimeinfra.NewAllocationRepository(root.DB())
	instanceRepo := instanceinfra.NewRepository(root.DB())
	defaultNodeName := defaultRuntimeNodeName(cfg)
	nodeSelector, nodeRepo, defaultNode, err := buildDefaultRuntimeNodeSelector(root, defaultNodeName)
	if err != nil {
		return nil, err
	}

	defaultNodeClient, err := buildDefaultNodeRuntimeClient(root, allocationRepo, defaultNode)
	if err != nil {
		return nil, err
	}
	executor := defaultNodeClient.executor
	checkerRunner := defaultNodeClient.checkerRunner
	module := containerruntime.Build(containerruntime.Deps{
		Config:                    cfg,
		Logger:                    log,
		ProvisioningRepository:    allocationRepo,
		CleanupRepository:         allocationRepo,
		ProvisioningRuntime:       executor,
		CleanupRuntime:            executor,
		FileRuntime:               executor,
		ImageRuntime:              executor,
		ManagedContainerInventory: executor,
		ManagedContainerStats:     executor,
		InteractiveExecutor:       executor,
	})
	nodeRouter := newRuntimeNodeExecutionRouter(cfg, log.Named("runtime_node_router"), allocationRepo, indexRepo, nodeRepo, defaultNodeName)
	if nodeRouter != nil && defaultNode != nil && defaultNode.ID > 0 {
		nodeRouter.rememberClient(defaultNode.ID, defaultNodeClient)
	}
	if err := migrateLegacyInstanceACLHandles(root.Context(), aclMigrationRepo, nodeRouter, defaultNodeClient, log.Named("runtime_acl_migration")); err != nil {
		if nodeRouter != nil {
			_ = nodeRouter.Close(root.Context())
		} else if defaultNodeClient != nil {
			_ = defaultNodeClient.Close(root.Context())
		}
		return nil, err
	}

	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewBackgroundJob(job.Name, job.Start, job.Stop))
	}
	runtimeNodeHealth := registerRuntimeNodeHealthJob(root, cfg, nodeRepo, &runtimeNodeStatsProbe{
		router:   nodeRouter,
		fallback: defaultNodeClient,
	}, log.Named("runtime_node_health"))

	contestContainerFiles := module.ContainerFiles
	contestCheckerRunner := checkerRunner
	lifecycleCloser := runtimeLifecycleCloser(defaultNodeClient)
	if nodeRouter != nil {
		contestContainerFiles = nodeRouter
		contestCheckerRunner = nodeRouter
		lifecycleCloser = nodeRouter
	}

	return &ContainerRuntimeModule{
		ChallengeImageRuntime:   module.ImageRuntime,
		ChallengeRuntimeProbe:   newChallengeRuntimeProbeAdapter(module.CleanupService, module.ProvisioningService, runtimePublishedAccessHost(cfg)),
		OpsRuntimeQuery:         newOpsRuntimeQueryAdapter(instanceRepo),
		OpsRuntimeStatsProvider: newOpsRuntimeStatsProviderAdapter(module.RuntimeStatsProvider),
		ContestContainerFiles:   contestContainerFiles,
		ContestCheckerRunner:    contestCheckerRunner,
		RuntimeNodeSelector:     nodeSelector,
		LifecycleCloser:         lifecycleCloser,
		nodeRouter:              nodeRouter,
		runtime:                 module,
		runtimeNodeHealth:       runtimeNodeHealth,
	}, nil
}

type runtimeNodeStatsProbe struct {
	router   *runtimeNodeExecutionRouter
	fallback *nodeRuntimeClient
}

func (p *runtimeNodeStatsProbe) ListManagedContainerStats(ctx context.Context, node runtimeentity.RuntimeNode) ([]runtimeapp.ManagedContainerStat, error) {
	if p == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	if p.router != nil {
		client, _, err := p.router.clientForConcreteNode(ctx, &node)
		if err != nil {
			return nil, err
		}
		return client.ListManagedContainerStats(ctx)
	}
	if p.fallback != nil {
		return p.fallback.ListManagedContainerStats(ctx)
	}
	return nil, runtimeports.ErrRuntimeNodeUnavailable
}

func registerRuntimeNodeHealthJob(root *Root, cfg *config.Config, repo *containerruntimeinfra.RuntimeNodeRepository, probe runtimeapp.NodeHealthProbe, logger *zap.Logger) *runtimeapp.NodeHealthService {
	if root == nil || cfg == nil || repo == nil || probe == nil || !cfg.Container.RuntimeNodeHealth.Enabled {
		return nil
	}
	service := runtimeapp.NewNodeHealthService(repo, probe, runtimeapp.NodeHealthOptions{
		PollInterval:     cfg.Container.RuntimeNodeHealth.PollInterval,
		ProbeTimeout:     cfg.Container.RuntimeNodeHealth.ProbeTimeout,
		StaleAfter:       cfg.Container.RuntimeNodeHealth.StaleAfter,
		FailureThreshold: cfg.Container.RuntimeNodeHealth.FailureThreshold,
	}, logger)
	root.RegisterBackgroundJob(NewLoopBackgroundJob("runtime_node_health", service.Run))
	return service
}

func (m *ContainerRuntimeModule) SetRuntimeNodeOfflineHandler(handler runtimeapp.NodeOfflineHandler) {
	if m == nil || m.runtimeNodeHealth == nil {
		return
	}
	m.runtimeNodeHealth.SetOfflineHandler(handler)
}

func runtimeConfigOrDefault(cfg *config.Config) *config.Config {
	if cfg == nil {
		return &config.Config{}
	}
	return cfg
}

func runtimePublishedAccessHost(cfg *config.Config) string {
	cfg = runtimeConfigOrDefault(cfg)
	return runtimecontracts.ResolveRuntimePublishedAccessHost(cfg.Container.PublicHost, cfg.Container.AccessHost)
}

func buildRuntimeHostExecutor(root *Root) runtimeports.RuntimeHostExecutor {
	client, err := buildDefaultRuntimeNodeClient(root)
	if err != nil || client == nil {
		return nil
	}
	return client.executor
}

func buildDefaultRuntimeNodeSelector(root *Root, defaultNodeName string) (runtimeports.RuntimeNodeSelector, *containerruntimeinfra.RuntimeNodeRepository, *runtimeentity.RuntimeNode, error) {
	if root == nil || root.DB() == nil {
		return nil, nil, nil, nil
	}
	cfg := root.Config()
	if cfg == nil {
		cfg = &config.Config{}
	}

	repo := containerruntimeinfra.NewRuntimeNodeRepository(root.DB())
	if repo == nil {
		return nil, nil, nil, nil
	}
	spec := runtimecontracts.RuntimeNodeBootstrapSpec{
		Name:        defaultNodeName,
		Endpoint:    defaultRuntimeNodeEndpoint(cfg),
		TLSIdentity: defaultRuntimeNodeTLSIdentity(cfg),
		Schedulable: true,
	}
	node, err := repo.EnsureDefaultNode(root.Context(), spec)
	if err != nil {
		return nil, nil, nil, err
	}
	return containerruntimeinfra.NewDefaultRuntimeNodeSelector(repo, defaultNodeName, runtimeNodeHealthStaleThreshold(cfg)), repo, node, nil
}

func buildDefaultRuntimeNodeClient(root *Root) (*nodeRuntimeClient, error) {
	if root == nil {
		return nil, nil
	}
	allocationRepo := containerruntimeinfra.NewAllocationRepository(root.DB())
	_, _, defaultNode, err := buildDefaultRuntimeNodeSelector(root, defaultRuntimeNodeName(root.Config()))
	if err != nil {
		return nil, err
	}
	return buildDefaultNodeRuntimeClient(root, allocationRepo, defaultNode)
}

func buildDefaultNodeRuntimeClient(root *Root, allocationRepo runtimeNodeAllocationRepository, node *runtimeentity.RuntimeNode) (*nodeRuntimeClient, error) {
	if root == nil {
		return nil, nil
	}
	client, err := buildRuntimeNodeClient(root.Context(), root.Config(), root.Logger(), allocationRepo, node)
	if err != nil {
		return nil, err
	}
	typedClient, ok := client.(*nodeRuntimeClient)
	if !ok {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	return typedClient, nil
}

func defaultRuntimeNodeName(cfg *config.Config) string {
	if cfg != nil && cfg.RuntimeAgent.Enabled {
		return "agent-default"
	}
	return "local-default"
}

func defaultRuntimeNodeEndpoint(cfg *config.Config) string {
	if cfg != nil && cfg.RuntimeAgent.Enabled {
		return cfg.RuntimeAgent.Endpoint
	}
	return "local://docker"
}

func defaultRuntimeNodeTLSIdentity(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	return cfg.RuntimeAgent.ServerName
}

type chainedRuntimeLifecycleCloser struct {
	closers []runtimeLifecycleCloser
}

func chainRuntimeLifecycleClosers(closers ...runtimeLifecycleCloser) runtimeLifecycleCloser {
	items := make([]runtimeLifecycleCloser, 0, len(closers))
	for _, closer := range closers {
		if closer == nil {
			continue
		}
		items = append(items, closer)
	}
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		return items[0]
	}
	return &chainedRuntimeLifecycleCloser{closers: items}
}

func (c *chainedRuntimeLifecycleCloser) Close(ctx context.Context) error {
	if c == nil {
		return nil
	}
	var closeErr error
	for _, closer := range c.closers {
		if closer == nil {
			continue
		}
		if err := closer.Close(ctx); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}
