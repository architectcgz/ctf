package composition

import (
	"context"

	"ctf-platform/internal/config"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	opsports "ctf-platform/internal/module/ops/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/internal/module/runtime/infrastructure/agentclient"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	runtimemodule "ctf-platform/internal/module/runtime/runtime"
)

type runtimeLifecycleCloser interface {
	Close(ctx context.Context) error
}

var (
	dialRuntimeAgent          = agentclient.DialContext
	newLocalRuntimeHostRunner = runtimeinfra.NewLocalHostExecutor
	newLocalCheckerRunner     = contestinfra.NewLocalCheckerRunner
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

	nodeRouter *runtimeNodeExecutionRouter
	runtime    *runtimemodule.Module
}

type RuntimeModule = ContainerRuntimeModule

func BuildContainerRuntimeModule(root *Root) (*ContainerRuntimeModule, error) {
	cfg := runtimeConfigOrDefault(root.Config())
	log := root.Logger()
	indexRepo := runtimeinfra.NewContainerNodeIndexRepository(root.DB())
	aclMigrationRepo := runtimeinfra.NewACLMigrationStateRepository(root.DB())
	allocationRepo := runtimeinfra.NewAllocationRepository(root.DB())
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
	module := runtimemodule.Build(runtimemodule.Deps{
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

	contestContainerFiles := module.ContestContainerFiles
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
	}, nil
}

func BuildRuntimeModule(root *Root) (*RuntimeModule, error) {
	return BuildContainerRuntimeModule(root)
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

func buildDefaultRuntimeNodeSelector(root *Root, defaultNodeName string) (runtimeports.RuntimeNodeSelector, *runtimeinfra.RuntimeNodeRepository, *runtimeentity.RuntimeNode, error) {
	if root == nil || root.DB() == nil {
		return nil, nil, nil, nil
	}
	cfg := root.Config()
	if cfg == nil {
		cfg = &config.Config{}
	}

	repo := runtimeinfra.NewRuntimeNodeRepository(root.DB())
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
	return runtimeinfra.NewDefaultRuntimeNodeSelector(repo, defaultNodeName), repo, node, nil
}

func buildDefaultRuntimeNodeClient(root *Root) (*nodeRuntimeClient, error) {
	if root == nil {
		return nil, nil
	}
	allocationRepo := runtimeinfra.NewAllocationRepository(root.DB())
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
