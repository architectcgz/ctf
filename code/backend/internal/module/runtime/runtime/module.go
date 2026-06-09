package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
	instanceports "ctf-platform/internal/module/instance/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type BackgroundJob struct {
	Name  string
	Start func(context.Context) error
	Stop  func(context.Context) error
}

type Module struct {
	BackgroundJobs []BackgroundJob

	ImageRuntime          *runtimeapp.ImageRuntimeService
	RuntimeStatsProvider  runtimeports.ManagedContainerStatsReader
	ContestContainerFiles contestports.AWDContainerFileWriter
	ProvisioningService   *runtimecmd.ProvisioningService
	CleanupService        *runtimecmd.RuntimeCleanupService

	ProvisioningRuntime       runtimeports.ContainerProvisioningRuntime
	CleanupRuntime            runtimeports.ContainerCleanupRuntime
	FileRuntime               runtimeports.ContainerFileRuntime
	ManagedContainerInventory runtimeports.ManagedContainerInventory
	InteractiveExecutor       runtimeports.ContainerInteractiveExecutor
}

type Deps struct {
	Config                    *config.Config
	Logger                    *zap.Logger
	DB                        *gorm.DB
	Cache                     *redislib.Client
	ProvisioningRuntime       runtimeports.ContainerProvisioningRuntime
	CleanupRuntime            runtimeports.ContainerCleanupRuntime
	FileRuntime               runtimeports.ContainerFileRuntime
	ImageRuntime              runtimeports.ContainerImageRuntime
	ManagedContainerInventory runtimeports.ManagedContainerInventory
	ManagedContainerStats     runtimeports.ManagedContainerStatsReader
	InteractiveExecutor       runtimeports.ContainerInteractiveExecutor
}

type runtimeInstanceRepository interface {
	instanceports.InstanceLookupRepository
	instanceports.InstanceUserLookupRepository
	instanceports.InstanceAccessRepository
	instanceports.UserVisibleInstanceRepository
	instanceports.TeacherInstanceQueryRepository
	instanceports.InstanceExtendRepository
	instanceports.InstanceStatusRepository
	instanceports.ProxyTicketInstanceReader
}

type runtimeModuleDeps struct {
	input                 Deps
	repo                  runtimeInstanceRepository
	cleanupService        *runtimecmd.RuntimeCleanupService
	provisioningService   *runtimecmd.ProvisioningService
	containerStatsService *runtimeapp.ContainerStatsService
	imageRuntime          *runtimeapp.ImageRuntimeService
	containerFiles        contestports.AWDContainerFileWriter
	containerPublicHost   string
}

func Build(deps Deps) *Module {
	internalDeps := buildRuntimeModuleDeps(deps)
	observabilityDeps := buildRuntimeObservabilityDeps(internalDeps)
	contestDeps := buildRuntimeContestDeps(internalDeps)

	return &Module{
		BackgroundJobs:            buildBackgroundJobs(internalDeps),
		ImageRuntime:              internalDeps.imageRuntime,
		RuntimeStatsProvider:      observabilityDeps.statsProvider,
		ContestContainerFiles:     contestDeps.containerFiles,
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
	repo := runtimeinfra.NewRepository(deps.DB)
	cleanupService := runtimecmd.NewRuntimeCleanupService(deps.CleanupRuntime, repo, log.Named("runtime_cleanup_service"))
	provisioningService := runtimecmd.NewProvisioningService(repo, deps.ProvisioningRuntime, &cfg.Container, log.Named("runtime_provisioning_service"))
	var containerStatsService *runtimeapp.ContainerStatsService
	if deps.ManagedContainerStats != nil {
		containerStatsService = runtimeapp.NewContainerStatsService(deps.ManagedContainerStats)
	}

	return runtimeModuleDeps{
		input:                 deps,
		repo:                  repo,
		cleanupService:        cleanupService,
		provisioningService:   provisioningService,
		containerStatsService: containerStatsService,
		imageRuntime:          runtimeapp.NewImageRuntimeService(deps.ImageRuntime),
		containerFiles:        runtimeapp.NewContainerFileService(deps.FileRuntime, log.Named("runtime_container_file_service")),
		containerPublicHost:   runtimecontracts.ResolveRuntimePublishedAccessHost(cfg.Container.PublicHost, cfg.Container.AccessHost),
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

type runtimeContestDeps struct {
	containerFiles contestports.AWDContainerFileWriter
}

func buildRuntimeContestDeps(deps runtimeModuleDeps) runtimeContestDeps {
	return runtimeContestDeps{
		containerFiles: deps.containerFiles,
	}
}
