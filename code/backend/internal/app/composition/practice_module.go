package composition

import (
	"context"

	"ctf-platform/internal/model"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

type PracticeModule struct {
	BackgroundTasks BackgroundTaskCloser
	Handler         *practicehttp.Handler
}

type practiceModuleDeps struct {
	commandRepo *practiceinfra.Repository
	scoreRepo   practiceports.PracticeScoreRepository
	rankingRepo practiceports.PracticeRankingRepository
}

type practiceModuleExternalDeps struct {
	instanceRepo   *runtimeinfra.Repository
	runtimeService practiceports.RuntimeInstanceService
	challengeRepo  practiceRuntimeChallengeContract
	imageStore     practiceRuntimeImageStore
	assessment     practiceRuntimeAssessmentService
}

type practiceRuntimeChallengeContract interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

type practiceRuntimeImageStore interface {
	FindByID(ctx context.Context, id int64) (*model.Image, error)
}

type practiceRuntimeAssessmentService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

func BuildPracticeModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) *PracticeModule {
	deps := buildPracticeModuleDeps(root)
	externalDeps := buildPracticeModuleExternalDeps(challenge, runtime, assessment)
	service := buildPracticeHandler(root, deps, externalDeps)
	service.StartBackgroundTasks(root.Context())
	service.SetEventBus(root.Events)
	scoreQueryService := practiceqry.NewScoreService(deps.rankingRepo, root.Cache(), root.Logger().Named("practice_score_query_service"), &root.Config().Score)
	root.RegisterBackgroundJob(NewLoopBackgroundJob("practice_instance_scheduler", service.RunProvisioningLoop))

	return &PracticeModule{
		BackgroundTasks: service,
		Handler:         practicehttp.NewHandler(service, scoreQueryService),
	}
}

func buildPracticeModuleDeps(root *Root) practiceModuleDeps {
	repo := practiceinfra.NewRepository(root.DB())
	return practiceModuleDeps{
		commandRepo: repo,
		scoreRepo:   repo,
		rankingRepo: repo,
	}
}

func buildPracticeModuleExternalDeps(challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) practiceModuleExternalDeps {
	return practiceModuleExternalDeps{
		instanceRepo:   runtime.PracticeInstanceRepository,
		runtimeService: runtime.PracticeRuntimeService,
		challengeRepo:  challenge.Catalog,
		imageStore:     challenge.ImageStore,
		assessment:     assessment.ProfileService,
	}
}

func buildPracticeHandler(root *Root, deps practiceModuleDeps, externalDeps practiceModuleExternalDeps) *practicecmd.Service {
	cfg := root.Config()
	log := root.Logger()
	cache := root.Cache()
	scoreService := practicecmd.NewScoreService(deps.scoreRepo, cache, log.Named("score_service"), &cfg.Score)
	return practicecmd.NewService(
		deps.commandRepo,
		externalDeps.challengeRepo,
		externalDeps.imageStore,
		externalDeps.instanceRepo,
		externalDeps.runtimeService,
		scoreService,
		externalDeps.assessment,
		cache,
		cfg,
		log.Named("practice_service"),
	)
}
