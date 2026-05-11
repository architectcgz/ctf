package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeqry "ctf-platform/internal/module/runtime/application/queries"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type Engine interface {
	runtimeports.ContainerProvisioningRuntime
	runtimeports.ContainerCleanupRuntime
	runtimeports.ContainerFileRuntime
	runtimeports.ContainerImageRuntime
	runtimeports.ManagedContainerInventory
	runtimeports.ManagedContainerStatsReader
	runtimeports.ContainerInteractiveExecutor
}

type BackgroundJob struct {
	Name  string
	Start func(context.Context) error
	Stop  func(context.Context) error
}

type Module struct {
	BackgroundJobs []BackgroundJob

	PracticeInstanceRepository practiceInstanceRepository
	PracticeRuntimeService     practiceports.RuntimeInstanceService
	ChallengeImageRuntime      challengeports.ImageRuntime
	ChallengeRuntimeProbe      challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery            opsports.RuntimeQuery
	OpsRuntimeStatsProvider    opsports.RuntimeStatsProvider
	ContestContainerFiles      contestports.AWDContainerFileWriter

	Engine Engine
}

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Cache  *redislib.Client
	Engine Engine
}

type runtimeInstanceRepository interface {
	runtimeports.InstanceLookupRepository
	runtimeports.InstanceUserLookupRepository
	runtimeports.InstanceAccessRepository
	runtimeports.UserVisibleInstanceRepository
	runtimeports.TeacherInstanceQueryRepository
	runtimeports.InstanceExtendRepository
	runtimeports.InstanceStatusRepository
	runtimeports.ProxyTicketInstanceReader
	runtimeports.CountRunningRepository
}

type practiceInstanceRepository interface {
	practiceports.PracticeInstanceLookupRepository
	practiceports.PracticeInstanceRuntimeWriteRepository
	practiceports.PracticeInstanceAWDOperationRepository
	practiceports.PracticeInstanceStatusRepository
	practiceports.PracticePendingInstanceRepository
	practiceports.PracticeInstanceStatsRepository
}

type runtimeModuleDeps struct {
	input                 Deps
	repo                  runtimeInstanceRepository
	practiceInstanceRepo  practiceInstanceRepository
	countRunningQuery     opsports.RuntimeQuery
	cleanupService        *runtimecmd.RuntimeCleanupService
	provisioningService   *runtimecmd.ProvisioningService
	containerStatsService *runtimeapp.ContainerStatsService
	imageRuntime          challengeports.ImageRuntime
	containerFiles        contestports.AWDContainerFileWriter
	containerPublicHost   string
}

func Build(deps Deps) *Module {
	internalDeps := buildRuntimeModuleDeps(deps)
	practiceDeps := buildRuntimePracticeDeps(internalDeps)
	challengeDeps := buildRuntimeChallengeDeps(internalDeps)
	opsDeps := buildRuntimeOpsDeps(internalDeps)
	contestDeps := buildRuntimeContestDeps(internalDeps)

	return &Module{
		BackgroundJobs:             buildBackgroundJobs(internalDeps),
		PracticeInstanceRepository: practiceDeps.instanceRepository,
		PracticeRuntimeService:     practiceDeps.runtimeService,
		ChallengeImageRuntime:      challengeDeps.imageRuntime,
		ChallengeRuntimeProbe:      challengeDeps.runtimeProbe,
		OpsRuntimeQuery:            opsDeps.query,
		OpsRuntimeStatsProvider:    opsDeps.statsProvider,
		ContestContainerFiles:      contestDeps.containerFiles,
		Engine:                     deps.Engine,
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
	cleanupService := runtimecmd.NewRuntimeCleanupService(deps.Engine, repo, log.Named("runtime_cleanup_service"))
	provisioningService := runtimecmd.NewProvisioningService(repo, deps.Engine, &cfg.Container, log.Named("runtime_provisioning_service"))
	var containerStatsService *runtimeapp.ContainerStatsService
	if deps.Engine != nil {
		containerStatsService = runtimeapp.NewContainerStatsService(deps.Engine)
	}

	return runtimeModuleDeps{
		input:                 deps,
		repo:                  repo,
		practiceInstanceRepo:  repo,
		countRunningQuery:     runtimeqry.NewCountRunningService(repo),
		cleanupService:        cleanupService,
		provisioningService:   provisioningService,
		containerStatsService: containerStatsService,
		imageRuntime:          runtimeapp.NewImageRuntimeService(deps.Engine),
		containerFiles:        runtimeapp.NewContainerFileService(deps.Engine, log.Named("runtime_container_file_service")),
		containerPublicHost:   cfg.Container.PublicHost,
	}
}

func buildBackgroundJobs(deps runtimeModuleDeps) []BackgroundJob {
	_ = deps
	return nil
}

type runtimePracticeDeps struct {
	instanceRepository practiceInstanceRepository
	runtimeService     practiceports.RuntimeInstanceService
}

func buildRuntimePracticeDeps(deps runtimeModuleDeps) runtimePracticeDeps {
	return runtimePracticeDeps{
		instanceRepository: deps.practiceInstanceRepo,
		runtimeService:     newRuntimePracticeServiceAdapter(deps.cleanupService, deps.provisioningService, deps.input.Engine),
	}
}

type runtimeChallengeDeps struct {
	imageRuntime challengeports.ImageRuntime
	runtimeProbe challengeports.ChallengeRuntimeProbe
}

func buildRuntimeChallengeDeps(deps runtimeModuleDeps) runtimeChallengeDeps {
	return runtimeChallengeDeps{
		imageRuntime: deps.imageRuntime,
		runtimeProbe: newRuntimeChallengeServiceAdapter(deps.cleanupService, deps.provisioningService, deps.containerPublicHost),
	}
}

type runtimeOpsDeps struct {
	query         opsports.RuntimeQuery
	statsProvider opsports.RuntimeStatsProvider
}

func buildRuntimeOpsDeps(deps runtimeModuleDeps) runtimeOpsDeps {
	return runtimeOpsDeps{
		query:         deps.countRunningQuery,
		statsProvider: newRuntimeOpsStatsProvider(deps.containerStatsService),
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
