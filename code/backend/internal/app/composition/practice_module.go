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
	commandRepo    practiceports.PracticeCommandRepository
	scoreRepo      practiceports.PracticeScoreRepository
	rankingRepo    practiceports.PracticeRankingRepository
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
	cfg := root.Config()
	log := root.Logger()
	cache := root.Cache()
	deps := buildPracticeModuleDeps(root, challenge, runtime, assessment)
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
	service.SetEventBus(root.Events)

	return &PracticeModule{
		BackgroundCloser: service,
		Handler:          practicehttp.NewHandler(service),
	}
}

func buildPracticeModuleDeps(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) practiceModuleDeps {
	repo := practiceinfra.NewRepository(root.DB())
	return practiceModuleDeps{
		commandRepo:    repo,
		scoreRepo:      repo,
		rankingRepo:    repo,
		instanceRepo:   runtime.practice.instanceRepository,
		runtimeService: runtime.practice.runtimeService,
		challengeRepo:  challenge.Catalog,
		imageStore:     challenge.ImageStore,
		assessment:     assessment.ProfileService,
	}
}
