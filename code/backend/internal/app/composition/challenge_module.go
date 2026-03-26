package composition

import (
	"context"

	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
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

type challengeModuleDeps struct {
	catalog              challengecontracts.ChallengeContract
	imageStore           challengecontracts.ImageStore
	imageRepo            challengeports.ImageRepository
	challengeCommandRepo challengeports.ChallengeCommandRepository
	challengeQueryRepo   challengeports.ChallengeQueryRepository
	flagRepo             challengeports.ChallengeFlagRepository
	imageUsageRepo       challengeports.ChallengeImageUsageRepository
	topologyRepo         challengeports.ChallengeTopologyRepository
	writeupRepo          challengeports.ChallengeWriteupRepository
	templateRepo         challengeports.EnvironmentTemplateRepository
	imageRuntime         challengeports.ImageRuntime
}

func BuildChallengeModule(root *Root, runtime *RuntimeModule) (*ChallengeModule, error) {
	cfg := root.Config()
	log := root.Logger()
	cache := root.Cache()
	deps := buildChallengeModuleDeps(root, runtime)

	imageCommandService := challengecmd.NewImageService(deps.imageRepo, deps.imageUsageRepo, deps.imageRuntime, log.Named("image_service"))
	imageQueryService := challengeqry.NewImageService(deps.imageRepo, cfg)
	challengeCommandService := challengecmd.NewChallengeService(deps.challengeCommandRepo, deps.imageRepo)
	challengeQueryService := challengeqry.NewChallengeService(deps.challengeQueryRepo, cache, &challengeqry.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, log.Named("challenge_service"))
	writeupCommandService := challengecmd.NewWriteupService(deps.writeupRepo)
	writeupQueryService := challengeqry.NewWriteupService(deps.writeupRepo)
	topologyCommandService := challengecmd.NewTopologyService(deps.topologyRepo, deps.templateRepo, deps.imageRepo)
	topologyQueryService := challengeqry.NewTopologyService(deps.topologyRepo, deps.templateRepo)
	flagCommandService, err := challengecmd.NewFlagService(deps.flagRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}
	flagQueryService, err := challengeqry.NewFlagService(deps.flagRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}

	return &ChallengeModule{
		BackgroundCloser: imageCommandService,
		Catalog:          deps.catalog,
		FlagHandler:      challengehttp.NewFlagHandler(flagCommandService, flagQueryService),
		FlagValidator:    flagQueryService,
		Handler:          challengehttp.NewHandler(challengeCommandService, challengeQueryService),
		ImageHandler:     challengehttp.NewImageHandler(imageCommandService, imageQueryService),
		ImageStore:       deps.imageStore,
		TopologyHandler:  challengehttp.NewTopologyHandler(topologyCommandService, topologyQueryService),
		WriteupHandler:   challengehttp.NewWriteupHandler(writeupCommandService, writeupQueryService),
	}, nil
}

func buildChallengeModuleDeps(root *Root, runtime *RuntimeModule) challengeModuleDeps {
	db := root.DB()

	challengeRepo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)

	return challengeModuleDeps{
		catalog:              challengeRepo,
		imageStore:           imageRepo,
		imageRepo:            imageRepo,
		challengeCommandRepo: challengeRepo,
		challengeQueryRepo:   challengeRepo,
		flagRepo:             challengeRepo,
		imageUsageRepo:       challengeRepo,
		topologyRepo:         challengeRepo,
		writeupRepo:          challengeRepo,
		templateRepo:         challengeinfra.NewTemplateRepository(db),
		imageRuntime:         runtime.challenge.imageRuntime,
	}
}
