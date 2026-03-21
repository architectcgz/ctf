package composition

import practiceModule "ctf-platform/internal/module/practice"

type PracticeModule struct {
	Handler *practiceModule.Handler
	Service *practiceModule.Service
}

func BuildPracticeModule(root *Root, challenge *ChallengeModule, container *ContainerModule, assessment *AssessmentModule) *PracticeModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := practiceModule.NewRepository(db)
	scoreService := practiceModule.NewScoreService(repo, cache, log.Named("score_service"), &cfg.Score)
	service := practiceModule.NewService(
		repo,
		challenge.Repository,
		challenge.ImageRepository,
		container.Repository,
		container.Service,
		scoreService,
		assessment.Service,
		cache,
		cfg,
		log.Named("practice_service"),
	)

	return &PracticeModule{
		Handler: practiceModule.NewHandler(service),
		Service: service,
	}
}
