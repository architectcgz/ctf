package runtime

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimeqry "ctf-platform/internal/module/container_runtime/application/queries"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

type BackgroundJob struct {
	Name  string
	Start func(context.Context) error
	Stop  func(context.Context) error
}

type Module struct {
	BackgroundJobs []BackgroundJob

	ImageRuntime         *runtimecmd.ImageRuntimeService
	RuntimeStatsProvider runtimeports.ManagedContainerStatsReader
	ContainerFiles       runtimeports.ContainerFileWriter
	ProvisioningService  *runtimecmd.ProvisioningService
	CleanupService       *runtimecmd.RuntimeCleanupService

	ProvisioningRuntime       runtimeports.ContainerProvisioningRuntime
	CleanupRuntime            runtimeports.ContainerCleanupRuntime
	FileRuntime               runtimeports.ContainerFileRuntime
	ManagedContainerInventory runtimeports.ManagedContainerInventory
	InteractiveExecutor       runtimeports.ContainerInteractiveExecutor
}

type Deps struct {
	Config                    *config.Config
	Logger                    *zap.Logger
	ProvisioningRepository    runtimecmd.ProvisioningRepository
	CleanupRepository         runtimecmd.RuntimeCleanupRepository
	ProvisioningRuntime       runtimeports.ContainerProvisioningRuntime
	CleanupRuntime            runtimeports.ContainerCleanupRuntime
	FileRuntime               runtimeports.ContainerFileRuntime
	ImageRuntime              runtimeports.ContainerImageRuntime
	ManagedContainerInventory runtimeports.ManagedContainerInventory
	ManagedContainerStats     runtimeports.ManagedContainerStatsReader
	InteractiveExecutor       runtimeports.ContainerInteractiveExecutor
}

type runtimeModuleDeps struct {
	input                 Deps
	cleanupService        *runtimecmd.RuntimeCleanupService
	provisioningService   *runtimecmd.ProvisioningService
	containerStatsService *runtimeqry.ContainerStatsService
	imageRuntime          *runtimecmd.ImageRuntimeService
	containerFiles        runtimeports.ContainerFileWriter
}

func Build(deps Deps) *Module {
	internalDeps := buildRuntimeModuleDeps(deps)
	observabilityDeps := buildRuntimeObservabilityDeps(internalDeps)

	return &Module{
		BackgroundJobs:            buildBackgroundJobs(internalDeps),
		ImageRuntime:              internalDeps.imageRuntime,
		RuntimeStatsProvider:      observabilityDeps.statsProvider,
		ContainerFiles:            internalDeps.containerFiles,
		ProvisioningService:       internalDeps.provisioningService,
		CleanupService:            internalDeps.cleanupService,
		ProvisioningRuntime:       deps.ProvisioningRuntime,
		CleanupRuntime:            deps.CleanupRuntime,
		FileRuntime:               deps.FileRuntime,
		ManagedContainerInventory: deps.ManagedContainerInventory,
		InteractiveExecutor:       deps.InteractiveExecutor,
	}
}

func buildRuntimeModuleDeps(deps Deps) runtimeModuleDeps {
	cfg := deps.Config
	log := deps.Logger
	if cfg == nil {
		cfg = &config.Config{}
	}
	if log == nil {
		log = zap.NewNop()
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(deps.CleanupRuntime, deps.CleanupRepository, log.Named("container_runtime_cleanup_service"))
	provisioningService := runtimecmd.NewProvisioningService(deps.ProvisioningRepository, deps.ProvisioningRuntime, &cfg.Container, log.Named("container_runtime_provisioning_service"))
	var containerStatsService *runtimeqry.ContainerStatsService
	if deps.ManagedContainerStats != nil {
		containerStatsService = runtimeqry.NewContainerStatsService(deps.ManagedContainerStats)
	}

	return runtimeModuleDeps{
		input:                 deps,
		cleanupService:        cleanupService,
		provisioningService:   provisioningService,
		containerStatsService: containerStatsService,
		imageRuntime:          runtimecmd.NewImageRuntimeService(deps.ImageRuntime),
		containerFiles:        runtimecmd.NewContainerFileService(deps.FileRuntime, log.Named("container_runtime_file_service")),
	}
}

func buildBackgroundJobs(deps runtimeModuleDeps) []BackgroundJob {
	_ = deps
	return nil
}

type runtimeObservabilityDeps struct {
	statsProvider runtimeports.ManagedContainerStatsReader
}

func buildRuntimeObservabilityDeps(deps runtimeModuleDeps) runtimeObservabilityDeps {
	return runtimeObservabilityDeps{
		statsProvider: deps.containerStatsService,
	}
}
