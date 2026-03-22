package composition

import challengeModule "ctf-platform/internal/module/challenge"

type ChallengeModule struct {
	FlagHandler     *challengeModule.FlagHandler
	FlagService     challengeModule.FlagValidator
	Handler         *challengeModule.Handler
	ImageHandler    *challengeModule.ImageHandler
	ImageRepository *challengeModule.ImageRepository
	ImageService    *challengeModule.ImageService
	Repository      challengeModule.ChallengeContract
	TopologyHandler *challengeModule.TopologyHandler
	WriteupHandler  *challengeModule.WriteupHandler
}

func BuildChallengeModule(root *Root, runtime *RuntimeModule) (*ChallengeModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	challengeRepo := challengeModule.NewRepository(db)
	imageRepo := challengeModule.NewImageRepository(db)
	imageService := challengeModule.NewImageService(imageRepo, challengeRepo, runtime.imageRuntimeService, cfg, log.Named("image_service"))
	challengeService := challengeModule.NewService(
		challengeRepo,
		imageRepo,
		cache,
		&challengeModule.Config{SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL},
		log.Named("challenge_service"),
	)
	writeupService := challengeModule.NewWriteupService(challengeRepo)
	templateRepo := challengeModule.NewTemplateRepository(db)
	topologyService := challengeModule.NewTopologyService(challengeRepo, templateRepo, imageRepo)
	flagService, err := challengeModule.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}

	return &ChallengeModule{
		FlagHandler:     challengeModule.NewFlagHandler(flagService),
		FlagService:     flagService,
		Handler:         challengeModule.NewHandler(challengeService),
		ImageHandler:    challengeModule.NewImageHandler(imageService),
		ImageRepository: imageRepo,
		ImageService:    imageService,
		Repository:      challengeRepo,
		TopologyHandler: challengeModule.NewTopologyHandler(topologyService),
		WriteupHandler:  challengeModule.NewWriteupHandler(writeupService),
	}, nil
}
