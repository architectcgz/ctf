package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
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
	BackgroundJobs  []BackgroundJob
	BackgroundTasks BackgroundTaskCloser
	Handler         *practicehttp.Handler
}

type Deps struct {
	AppContext context.Context
	Config     *config.Config
	Logger     *zap.Logger
	DB         *gorm.DB
	Cache      *redislib.Client
	Events     platformevents.Bus
	// InstanceRepo   practiceports.InstanceRepository
	InstanceRepo interface {
		practiceports.PracticeInstanceLookupRepository
		practiceports.PracticeInstanceRuntimeWriteRepository
		practiceports.PracticeInstanceAWDOperationRepository
		practiceports.PracticeInstanceStatusRepository
		practiceports.PracticePendingInstanceRepository
		practiceports.PracticeInstanceStatsRepository
	}
	RuntimeService practiceports.RuntimeInstanceService
	ChallengeRepo  challengecontracts.PracticeChallengeContract
	ImageStore     challengecontracts.ImageStore
	Assessment     assessmentcontracts.ProfileService
}

type moduleDeps struct {
	input Deps
	// commandRepo    practiceports.PracticeCommandRepository
	commandRepo interface {
		practiceports.PracticeInstanceStartTxManager
		practiceports.PracticeInstanceRestartTxManager
		practiceports.PracticeAWDServiceOperationTxManager
		practiceports.PracticeInstanceStartTxRepository
		practiceports.PracticeInstanceRestartTxRepository
		practiceports.PracticeAWDServiceOperationTxRepository
		practiceports.PracticeContestLookupRepository
		practiceports.PracticeContestChallengeLookupRepository
		practiceports.PracticeContestAWDServiceRepository
		practiceports.PracticeContestAWDInstanceRepository
		practiceports.PracticeContestTeamRepository
		practiceports.PracticeContestRegistrationRepository
		practiceports.PracticeSubmissionWriteRepository
		practiceports.PracticeSolvedSubmissionRepository
		practiceports.PracticeSubmissionHistoryRepository
		practiceports.PracticeSubmissionConstraintRepository
		practiceports.PracticeUserLookupRepository
		practiceports.PracticeManualReviewListRepository
		practiceports.PracticeManualReviewLookupRepository
	}
	// scoreRepo      practiceports.PracticeScoreRepository
	scoreRepo interface {
		practiceports.PracticeChallengeScoreRepository
		practiceports.PracticeSolvedChallengeRepository
		practiceports.PracticeUserScoreWriteRepository
	}
	// rankingRepo    practiceports.PracticeRankingRepository
	rankingRepo interface {
		practiceports.PracticeUserScoreReadRepository
		practiceports.PracticeRankingListRepository
		practiceports.PracticeUserDirectoryRepository
	}
	// instanceRepo   practiceports.InstanceRepository
	instanceRepo interface {
		practiceports.PracticeInstanceLookupRepository
		practiceports.PracticeInstanceRuntimeWriteRepository
		practiceports.PracticeInstanceAWDOperationRepository
		practiceports.PracticeInstanceStatusRepository
		practiceports.PracticePendingInstanceRepository
		practiceports.PracticeInstanceStatsRepository
	}
	runtimeService practiceports.RuntimeInstanceService
	challengeRepo  challengecontracts.PracticeChallengeContract
	imageStore     challengecontracts.ImageStore
	assessment     assessmentcontracts.ProfileService
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	service, rankingService := buildHandler(internalDeps)
	service.StartBackgroundTasks(deps.AppContext)
	service.SetEventBus(deps.Events)

	return &Module{
		BackgroundJobs: []BackgroundJob{
			{Name: "practice_instance_scheduler", Run: service.RunProvisioningLoop},
		},
		BackgroundTasks: service,
		Handler:         practicehttp.NewHandler(service, rankingService),
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	repo := practiceinfra.NewRepository(deps.DB)
	return moduleDeps{
		input:          deps,
		commandRepo:    repo,
		scoreRepo:      repo,
		rankingRepo:    repo,
		instanceRepo:   deps.InstanceRepo,
		runtimeService: deps.RuntimeService,
		challengeRepo:  deps.ChallengeRepo,
		imageStore:     deps.ImageStore,
		assessment:     deps.Assessment,
	}
}

func buildHandler(deps moduleDeps) (*practicecmd.Service, *practiceqry.ScoreService) {
	cfg := deps.input.Config
	log := deps.input.Logger
	cache := deps.input.Cache

	scoreService := practicecmd.NewScoreService(deps.scoreRepo, cache, log.Named("score_service"), &cfg.Score)
	service := practicecmd.NewService(
		deps.commandRepo,
		deps.challengeRepo,
		deps.imageStore,
		deps.instanceRepo,
		deps.runtimeService,
		scoreService,
		deps.assessment,
		cache,
		cfg,
		log.Named("practice_service"),
	)
	rankingService := practiceqry.NewScoreService(deps.rankingRepo, cache, log.Named("practice_score_query_service"), &cfg.Score)

	return service, rankingService
}
