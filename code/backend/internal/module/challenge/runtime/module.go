package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
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
	AppContext   context.Context
	Config       *config.Config
	Logger       *zap.Logger
	DB           *gorm.DB
	Cache        *redislib.Client
	Events       platformevents.Bus
	ImageRuntime challengeports.ImageRuntime
	RuntimeProbe challengeports.ChallengeRuntimeProbe
}

type moduleDeps struct {
	input      Deps
	catalog    challengecontracts.ChallengeContract
	imageStore challengecontracts.ImageStore
	rawRepo    *challengeinfra.Repository
	// imageRepo               challengeports.ImageRepository
	imageRepo interface {
		challengeports.ImageCommandRepository
		challengeports.ImageQueryRepository
		challengeports.ImageBuildJobRepository
	}
	// challengeCommandRepo    challengeports.ChallengeCommandRepository
	challengeCommandRepo interface {
		CreateWithHints(ctx context.Context, challenge *challengeentity.Challenge, hints []*challengeentity.ChallengeHint) error
		FindByID(ctx context.Context, id int64) (*challengeentity.Challenge, error)
		Update(ctx context.Context, challenge *challengeentity.Challenge) error
		UpdateWithHints(ctx context.Context, challenge *challengeentity.Challenge, hints []*challengeentity.ChallengeHint, replaceHints bool) error
		Delete(ctx context.Context, id int64) error
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
	imageBuildService := buildImageBuildService(internalDeps)
	publishCheckService, coreHandler := buildCoreHandler(internalDeps, imageBuildService)
	flagHandler, flagValidator, err := buildFlagHandler(internalDeps)
	if err != nil {
		return nil, err
	}
	if imageBuildService != nil && deps.Config != nil && deps.Config.Container.Registry.BuildEnabled {
		imageBuildService.StartBackgroundTasks(deps.AppContext)
	}

	module := &Module{
		BackgroundTasks:     backgroundTaskGroup{imageCommandService, imageBuildService},
		AWDChallengeHandler: buildAWDChallengeHandler(internalDeps, imageBuildService),
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
			Run:  publishCheckService.RunPublishCheckLoop,
		})
	}
	return module, nil
}

func newModuleDeps(deps Deps) moduleDeps {
	challengeRepo := challengeinfra.NewRepository(deps.DB)
	imageRepo := challengeinfra.NewImageRepository(deps.DB)
	flagRepo := challengeinfra.NewFlagRepository(challengeRepo)
	awdChallengeRepo := challengeinfra.NewAWDChallengeRepository(challengeRepo)
	writeupRepo := challengeinfra.NewWriteupServiceRepository(challengeRepo)

	return moduleDeps{
		input:                   deps,
		catalog:                 challengeinfra.NewContractRepository(challengeRepo),
		imageStore:              imageRepo,
		rawRepo:                 challengeRepo,
		imageRepo:               imageRepo,
		challengeCommandRepo:    challengeRepo,
		challengeQueryRepo:      challengeinfra.NewChallengeQueryRepository(challengeRepo),
		awdChallengeCommandRepo: awdChallengeRepo,
		awdChallengeQueryRepo:   awdChallengeRepo,
		flagRepo:                flagRepo,
		imageUsageRepo:          challengeRepo,
		topologyRepo:            challengeinfra.NewTopologyServiceRepository(challengeRepo),
		writeupRepo:             writeupRepo,
		templateRepo:            challengeinfra.NewTemplateRepository(deps.DB),
		imageRuntime:            deps.ImageRuntime,
		runtimeProbe:            deps.RuntimeProbe,
	}
}
