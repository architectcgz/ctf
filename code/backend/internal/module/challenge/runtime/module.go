package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type BackgroundJob struct {
	Name string
	Run  func(context.Context)
}

type BackgroundTaskCloser interface {
	Close(ctx context.Context) error
}

type Module struct {
	BackgroundJobs      []BackgroundJob
	BackgroundTasks     BackgroundTaskCloser
	AWDChallengeHandler *challengehttp.AWDChallengeHandler
	AWDChallengeQuery   challengeports.AWDChallengeQueryRepository
	Catalog             challengecontracts.ChallengeContract
	FlagHandler         *challengehttp.FlagHandler
	FlagValidator       challengecontracts.FlagValidator
	Handler             *challengehttp.Handler
	ImageHandler        *challengehttp.ImageHandler
	ImageStore          challengecontracts.ImageStore
	TopologyHandler     *challengehttp.TopologyHandler
	WriteupHandler      *challengehttp.WriteupHandler
}

type Deps struct {
	AppContext    context.Context
	Config        *config.Config
	Logger        *zap.Logger
	DB            *gorm.DB
	Cache         *redislib.Client
	ImageRuntime  challengeports.ImageRuntime
	RuntimeProbe  challengeports.ChallengeRuntimeProbe
	Notifications challengecmd.ChallengeNotificationSender
}

type moduleDeps struct {
	input      Deps
	catalog    challengecontracts.ChallengeContract
	imageStore challengecontracts.ImageStore
	// imageRepo               challengeports.ImageRepository
	imageRepo interface {
		challengeports.ImageCommandRepository
		challengeports.ImageQueryRepository
		challengeports.ImageBuildJobRepository
	}
	// challengeCommandRepo    challengeports.ChallengeCommandRepository
	challengeCommandRepo interface {
		challengeports.ChallengeWriteRepository
		challengeports.ChallengeInstanceUsageRepository
		challengeports.ChallengePublishCheckRepository
		challengeports.ChallengePackageRevisionRepository
	}
	// challengeQueryRepo      challengeports.ChallengeQueryRepository
	challengeQueryRepo interface {
		challengeports.ChallengeReadRepository
		challengeports.ChallengePublishedRepository
		challengeports.ChallengeStatsRepository
		challengeports.ChallengeBatchStatsRepository
	}
	awdChallengeCommandRepo challengeports.AWDChallengeCommandRepository
	awdChallengeQueryRepo   challengeports.AWDChallengeQueryRepository
	flagRepo                challengeports.ChallengeFlagRepository
	imageUsageRepo          challengeports.ChallengeImageUsageRepository
	// topologyRepo            challengeports.ChallengeTopologyRepository
	topologyRepo interface {
		challengeports.ChallengeTopologyChallengeLookupRepository
		challengeports.ChallengeTopologyReadRepository
		challengeports.ChallengeTopologyWriteRepository
	}
	// writeupRepo             challengeports.ChallengeWriteupRepository
	writeupRepo interface {
		challengeports.ChallengeWriteupChallengeLookupRepository
		challengeports.ChallengeWriteupUserLookupRepository
		challengeports.ChallengeAdminWriteupRepository
		challengeports.ChallengeReleasedWriteupRepository
		challengeports.ChallengeWriteupSolveStatusRepository
		challengeports.ChallengeSubmissionWriteupRepository
		challengeports.ChallengeTeacherSubmissionWriteupRepository
		challengeports.ChallengeSolutionQueryRepository
	}
	// templateRepo            challengeports.EnvironmentTemplateRepository
	templateRepo interface {
		challengeports.EnvironmentTemplateCommandRepository
		challengeports.EnvironmentTemplateQueryRepository
		challengeports.EnvironmentTemplateUsageRepository
	}
	imageRuntime challengeports.ImageRuntime
	runtimeProbe challengeports.ChallengeRuntimeProbe
}

func Build(deps Deps) (*Module, error) {
	internalDeps := newModuleDeps(deps)

	imageCommandService, imageHandler := buildImageHandler(internalDeps)
	coreService, coreHandler := buildCoreHandler(internalDeps)
	flagHandler, flagValidator, err := buildFlagHandler(internalDeps)
	if err != nil {
		return nil, err
	}

	module := &Module{
		BackgroundTasks:     imageCommandService,
		AWDChallengeHandler: buildAWDChallengeHandler(internalDeps),
		AWDChallengeQuery:   internalDeps.awdChallengeQueryRepo,
		Catalog:             internalDeps.catalog,
		FlagHandler:         flagHandler,
		FlagValidator:       flagValidator,
		Handler:             coreHandler,
		ImageHandler:        imageHandler,
		ImageStore:          internalDeps.imageStore,
		TopologyHandler:     buildTopologyHandler(internalDeps),
		WriteupHandler:      buildWriteupHandler(internalDeps),
	}
	if deps.Config != nil && deps.Config.Challenge.PublishCheck.Enabled {
		module.BackgroundJobs = append(module.BackgroundJobs, BackgroundJob{
			Name: "challenge_publish_check_worker",
			Run:  coreService.RunPublishCheckLoop,
		})
	}
	return module, nil
}

func newModuleDeps(deps Deps) moduleDeps {
	challengeRepo := challengeinfra.NewRepository(deps.DB)
	imageRepo := challengeinfra.NewImageRepository(deps.DB)

	return moduleDeps{
		input:                   deps,
		catalog:                 challengeRepo,
		imageStore:              imageRepo,
		imageRepo:               imageRepo,
		challengeCommandRepo:    challengeRepo,
		challengeQueryRepo:      challengeRepo,
		awdChallengeCommandRepo: challengeRepo,
		awdChallengeQueryRepo:   challengeRepo,
		flagRepo:                challengeRepo,
		imageUsageRepo:          challengeRepo,
		topologyRepo:            challengeRepo,
		writeupRepo:             challengeRepo,
		templateRepo:            challengeinfra.NewTemplateRepository(deps.DB),
		imageRuntime:            deps.ImageRuntime,
		runtimeProbe:            deps.RuntimeProbe,
	}
}

func buildAWDChallengeHandler(deps moduleDeps) *challengehttp.AWDChallengeHandler {
	commandService := challengecmd.NewAWDChallengeCommandFacade(deps.input.DB, deps.awdChallengeCommandRepo)
	queryService := challengeqry.NewAWDChallengeQueryService(deps.awdChallengeQueryRepo)
	return challengehttp.NewAWDChallengeHandler(commandService, queryService)
}

func buildImageHandler(deps moduleDeps) (*challengecmd.ImageService, *challengehttp.ImageHandler) {
	imageCommandService := challengecmd.NewImageService(
		deps.imageRepo,
		deps.imageUsageRepo,
		deps.imageRuntime,
		deps.input.Logger.Named("image_service"),
	)
	imageCommandService.StartBackgroundTasks(deps.input.AppContext)
	imageQueryService := challengeqry.NewImageService(deps.imageRepo, deps.input.Config)
	return imageCommandService, challengehttp.NewImageHandler(imageCommandService, imageQueryService)
}

func buildCoreHandler(deps moduleDeps) (*challengecmd.ChallengeService, *challengehttp.Handler) {
	cfg := deps.input.Config
	challengeCommandService := challengecmd.NewChallengeService(
		deps.input.DB,
		deps.challengeCommandRepo,
		deps.imageRepo,
		deps.topologyRepo,
		deps.challengeCommandRepo,
		deps.runtimeProbe,
		challengecmd.SelfCheckConfig{
			RuntimeCreateTimeout:     cfg.Container.CreateTimeout,
			FlagGlobalSecret:         cfg.Container.FlagGlobalSecret,
			PublishCheckPollInterval: cfg.Challenge.PublishCheck.PollInterval,
			PublishCheckBatchSize:    cfg.Challenge.PublishCheck.BatchSize,
		},
		deps.input.Logger.Named("challenge_command_service"),
		deps.input.Notifications,
	)
	challengeQueryService := challengeqry.NewChallengeService(deps.challengeQueryRepo, deps.input.Cache, &challengeqry.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, deps.input.Logger.Named("challenge_service"))
	return challengeCommandService, challengehttp.NewHandler(challengeCommandService, challengeQueryService)
}

func buildFlagHandler(deps moduleDeps) (*challengehttp.FlagHandler, challengecontracts.FlagValidator, error) {
	flagCommandService, err := challengecmd.NewFlagService(deps.flagRepo, deps.input.Config.Container.FlagGlobalSecret)
	if err != nil {
		return nil, nil, err
	}
	flagQueryService, err := challengeqry.NewFlagService(deps.flagRepo, deps.input.Config.Container.FlagGlobalSecret)
	if err != nil {
		return nil, nil, err
	}
	return challengehttp.NewFlagHandler(flagCommandService, flagQueryService), flagQueryService, nil
}

func buildTopologyHandler(deps moduleDeps) *challengehttp.TopologyHandler {
	topologyCommandService := challengecmd.NewTopologyService(deps.topologyRepo, deps.templateRepo, deps.imageRepo)
	topologyQueryService := challengeqry.NewTopologyService(deps.topologyRepo, deps.templateRepo)
	return challengehttp.NewTopologyHandler(topologyCommandService, topologyQueryService)
}

func buildWriteupHandler(deps moduleDeps) *challengehttp.WriteupHandler {
	writeupCommandService := challengecmd.NewWriteupService(deps.writeupRepo)
	writeupQueryService := challengeqry.NewWriteupService(deps.writeupRepo)
	return challengehttp.NewWriteupHandler(writeupCommandService, writeupQueryService)
}
