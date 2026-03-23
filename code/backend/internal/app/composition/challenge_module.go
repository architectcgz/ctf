package composition

import (
	"context"

	challengeModule "ctf-platform/internal/module/challenge"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

type asyncTaskCloser interface {
	Close(ctx context.Context) error
}

type ChallengeModule struct {
	BackgroundCloser asyncTaskCloser
	Catalog          challengecontracts.ChallengeContract
	FlagHandler      *challengeModule.FlagHandler
	FlagValidator    challengecontracts.FlagValidator
	Handler          *challengeModule.Handler
	ImageHandler     *challengeModule.ImageHandler
	ImageStore       challengecontracts.ImageStore
	TopologyHandler  *challengeModule.TopologyHandler
	WriteupHandler   *challengeModule.WriteupHandler
}

func BuildChallengeModule(root *Root, runtime *RuntimeModule) (*ChallengeModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	challengeRepo := challengeModule.NewRepository(db)
	imageRepo := challengeModule.NewImageRepository(db)
	imageService := challengeModule.NewImageService(imageRepo, challengeRepo, runtime.challenge.imageRuntime, cfg, log.Named("image_service"))
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
		BackgroundCloser: imageService,
		Catalog:          challengeRepo,
		FlagHandler:      challengeModule.NewFlagHandler(flagService),
		FlagValidator:    flagService,
		Handler:          challengeModule.NewHandler(challengeService),
		ImageHandler:     challengeModule.NewImageHandler(imageService),
		ImageStore:       imageRepo,
		TopologyHandler:  challengeModule.NewTopologyHandler(topologyService),
		WriteupHandler:   challengeModule.NewWriteupHandler(writeupService),
	}, nil
}
