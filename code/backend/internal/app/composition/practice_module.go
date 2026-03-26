package composition

import (
	"context"

	"ctf-platform/internal/model"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type PracticeModule struct {
	BackgroundCloser asyncTaskCloser
	Handler          *practicehttp.Handler
}

type practiceModuleDeps struct {
	commandRepo practiceports.PracticeCommandRepository
	scoreRepo   practiceports.PracticeScoreRepository
	rankingRepo practiceports.PracticeRankingRepository
}

type practiceModuleExternalDeps struct {
	instanceRepo   practiceports.InstanceRepository
	runtimeService practiceports.RuntimeInstanceService
	challengeRepo  practiceRuntimeChallengeContract
	imageStore     practiceRuntimeImageStore
	assessment     practiceRuntimeAssessmentService
}

type practiceRuntimeChallengeContract interface {
	FindByID(id int64) (*model.Challenge, error)
	FindHintByLevel(challengeID int64, level int) (*model.ChallengeHint, error)
	CreateHintUnlock(unlock *model.ChallengeHintUnlock) error
	FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error)
}

type practiceRuntimeImageStore interface {
	FindByID(id int64) (*model.Image, error)
}

type practiceRuntimeAssessmentService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

func BuildPracticeModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) *PracticeModule {
	deps := buildPracticeModuleDeps(root)
	externalDeps := buildPracticeModuleExternalDeps(challenge, runtime, assessment)
	service := buildPracticeHandler(root, deps, externalDeps)
	service.SetEventBus(root.Events)

	return &PracticeModule{
		BackgroundCloser: service,
		Handler:          practicehttp.NewHandler(service),
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
		instanceRepo:   runtime.practice.instanceRepository,
		runtimeService: runtime.practice.runtimeService,
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
