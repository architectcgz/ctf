package composition

import (
	"context"

	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
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
	imageCommandService := challengecmd.NewImageService(imageRepo, challengeRepo, runtime.challenge.imageRuntime, log.Named("image_service"))
	imageQueryService := challengeqry.NewImageService(imageRepo, cfg)
	challengeCommandService := challengecmd.NewChallengeService(challengeRepo, imageRepo)
	challengeQueryService := challengeqry.NewChallengeService(challengeRepo, cache, &challengeqry.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, log.Named("challenge_service"))
	writeupCommandService := challengecmd.NewWriteupService(challengeRepo)
	writeupQueryService := challengeqry.NewWriteupService(challengeRepo)
	templateRepo := challengeinfra.NewTemplateRepository(db)
	topologyCommandService := challengecmd.NewTopologyService(challengeRepo, templateRepo, imageRepo)
	topologyQueryService := challengeqry.NewTopologyService(challengeRepo, templateRepo)
	flagCommandService, err := challengecmd.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}
	flagQueryService, err := challengeqry.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}

	return &ChallengeModule{
		BackgroundCloser: imageCommandService,
		Catalog:          challengeRepo,
		FlagHandler:      challengehttp.NewFlagHandler(flagCommandService, flagQueryService),
		FlagValidator:    flagQueryService,
		Handler:          challengehttp.NewHandler(challengeCommandService, challengeQueryService),
		ImageHandler:     challengehttp.NewImageHandler(imageCommandService, imageQueryService),
		ImageStore:       imageRepo,
		TopologyHandler:  challengehttp.NewTopologyHandler(topologyCommandService, topologyQueryService),
		WriteupHandler:   challengehttp.NewWriteupHandler(writeupCommandService, writeupQueryService),
	}, nil
}
