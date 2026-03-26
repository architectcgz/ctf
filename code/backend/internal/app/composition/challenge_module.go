package composition

import (
	"context"

	"ctf-platform/internal/config"
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
	deps := buildChallengeModuleDeps(root, runtime)

	imageCommandService, imageHandler := buildChallengeImageHandler(root, deps)
	coreHandler := buildChallengeCoreHandler(root, deps)
	flagHandler, flagValidator, err := buildChallengeFlagHandler(cfg, deps)
	if err != nil {
		return nil, err
	}
	topologyHandler := buildChallengeTopologyHandler(deps)
	writeupHandler := buildChallengeWriteupHandler(deps)

	return &ChallengeModule{
		BackgroundCloser: imageCommandService,
		Catalog:          deps.catalog,
		FlagHandler:      flagHandler,
		FlagValidator:    flagValidator,
		Handler:          coreHandler,
		ImageHandler:     imageHandler,
		ImageStore:       deps.imageStore,
		TopologyHandler:  topologyHandler,
		WriteupHandler:   writeupHandler,
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
		imageRuntime:         runtime.ChallengeImageRuntime,
	}
}

func buildChallengeImageHandler(root *Root, deps challengeModuleDeps) (*challengecmd.ImageService, *challengehttp.ImageHandler) {
	imageCommandService := challengecmd.NewImageService(
		deps.imageRepo,
		deps.imageUsageRepo,
		deps.imageRuntime,
		root.Logger().Named("image_service"),
	)
	imageQueryService := challengeqry.NewImageService(deps.imageRepo, root.Config())
	return imageCommandService, challengehttp.NewImageHandler(imageCommandService, imageQueryService)
}

func buildChallengeCoreHandler(root *Root, deps challengeModuleDeps) *challengehttp.Handler {
	challengeCommandService := challengecmd.NewChallengeService(deps.challengeCommandRepo, deps.imageRepo)
	challengeQueryService := challengeqry.NewChallengeService(deps.challengeQueryRepo, root.Cache(), &challengeqry.Config{
		SolvedCountCacheTTL: root.Config().Challenge.SolvedCountCacheTTL,
	}, root.Logger().Named("challenge_service"))
	return challengehttp.NewHandler(challengeCommandService, challengeQueryService)
}

func buildChallengeFlagHandler(cfg *config.Config, deps challengeModuleDeps) (*challengehttp.FlagHandler, challengecontracts.FlagValidator, error) {
	flagCommandService, err := challengecmd.NewFlagService(deps.flagRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, nil, err
	}
	flagQueryService, err := challengeqry.NewFlagService(deps.flagRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, nil, err
	}
	return challengehttp.NewFlagHandler(flagCommandService, flagQueryService), flagQueryService, nil
}

func buildChallengeTopologyHandler(deps challengeModuleDeps) *challengehttp.TopologyHandler {
	topologyCommandService := challengecmd.NewTopologyService(deps.topologyRepo, deps.templateRepo, deps.imageRepo)
	topologyQueryService := challengeqry.NewTopologyService(deps.topologyRepo, deps.templateRepo)
	return challengehttp.NewTopologyHandler(topologyCommandService, topologyQueryService)
}

func buildChallengeWriteupHandler(deps challengeModuleDeps) *challengehttp.WriteupHandler {
	writeupCommandService := challengecmd.NewWriteupService(deps.writeupRepo)
	writeupQueryService := challengeqry.NewWriteupService(deps.writeupRepo)
	return challengehttp.NewWriteupHandler(writeupCommandService, writeupQueryService)
}
