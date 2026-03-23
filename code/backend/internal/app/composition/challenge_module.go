package composition

import (
	"context"

	challengemodule "ctf-platform/internal/module/challenge"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengeapp "ctf-platform/internal/module/challenge/application"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
)

type asyncTaskCloser interface {
	Close(ctx context.Context) error
}

type ChallengeModule struct {
	BackgroundCloser asyncTaskCloser
	Catalog          challengecontracts.ChallengeContract
	FlagHandler      *challengehttp.FlagHandler
	FlagValidator    challengecontracts.FlagValidator
	Handler          *challengehttp.Handler
	ImageHandler     *challengehttp.ImageHandler
	ImageStore       challengecontracts.ImageStore
	TopologyHandler  *challengehttp.TopologyHandler
	WriteupHandler   *challengehttp.WriteupHandler
}

func BuildChallengeModule(root *Root, runtime *RuntimeModule) (*ChallengeModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	challengeRepo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageService := challengeapp.NewImageService(imageRepo, challengeRepo, runtime.challenge.imageRuntime, cfg, log.Named("image_service"))
	challengeService := challengeapp.NewService(
		challengeRepo,
		imageRepo,
		cache,
		&challengemodule.Config{SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL},
		log.Named("challenge_service"),
	)
	writeupService := challengeapp.NewWriteupService(challengeRepo)
	templateRepo := challengeinfra.NewTemplateRepository(db)
	topologyService := challengeapp.NewTopologyService(challengeRepo, templateRepo, imageRepo)
	flagService, err := challengeapp.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}

	return &ChallengeModule{
		BackgroundCloser: imageService,
		Catalog:          challengeRepo,
		FlagHandler:      challengehttp.NewFlagHandler(flagService),
		FlagValidator:    flagService,
		Handler:          challengehttp.NewHandler(challengeService),
		ImageHandler:     challengehttp.NewImageHandler(imageService),
		ImageStore:       imageRepo,
		TopologyHandler:  challengehttp.NewTopologyHandler(topologyService),
		WriteupHandler:   challengehttp.NewWriteupHandler(writeupService),
	}, nil
}
