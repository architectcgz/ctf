package runtime

import (
	"context"
	"errors"

	"ctf-platform/internal/config"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecore "ctf-platform/internal/module/challenge/application/challengecore"
	challengeimport "ctf-platform/internal/module/challenge/application/challengeimport"
	challengepackageexport "ctf-platform/internal/module/challenge/application/challengepackageexport"
	challengepublishcheck "ctf-platform/internal/module/challenge/application/challengepublishcheck"
	challengeselfcheck "ctf-platform/internal/module/challenge/application/challengeselfcheck"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformsharedfs "ctf-platform/internal/platform/storage/sharedfs"
)

type backgroundTaskGroup []BackgroundTaskCloser

func (g backgroundTaskGroup) Close(ctx context.Context) error {
	var errs []error
	for _, closer := range g {
		if closer == nil {
			continue
		}
		if err := closer.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

var (
	imageBuildDockerBuilderFactory = func(registry config.ContainerRegistryConfig) challengeports.DockerImageBuilder {
		return challengeinfra.NewDockerCLIImageBuilderWithConfig(challengeinfra.DockerCLIImageBuilderConfig{
			RegistryServer: registry.Server,
			Username:       registry.Username,
			Password:       registry.Password,
			IdentityToken:  registry.IdentityToken,
		})
	}
	imageBuildRegistryVerifierFactory = func(registry config.ContainerRegistryConfig) challengeports.RegistryVerifier {
		return challengeinfra.NewRegistryClient(challengeinfra.RegistryClientConfig{
			Scheme:        registry.Scheme,
			Server:        registry.Server,
			AccessServer:  registry.AccessServer,
			Username:      registry.Username,
			Password:      registry.Password,
			IdentityToken: registry.IdentityToken,
		}, nil)
	}
)

func buildImageBuildService(deps moduleDeps) *challengecmd.ImageBuildService {
	cfg := deps.input.Config
	if cfg == nil {
		return nil
	}
	registry := cfg.Container.Registry
	if !registry.Enabled && !registry.BuildEnabled {
		return nil
	}

	options := []challengecmd.ImageBuildOption{
		challengecmd.WithImageBuildLogger(deps.input.Logger.Named("image_build_service")),
	}
	if builder := imageBuildDockerBuilderFactory(registry); builder != nil {
		options = append(options, challengecmd.WithImageBuildDockerBuilder(builder))
	}
	if verifier := imageBuildRegistryVerifierFactory(registry); verifier != nil {
		options = append(options, challengecmd.WithImageBuildRegistryVerifier(verifier))
	}

	return challengecmd.NewImageBuildService(
		challengeinfra.NewImageBuildRepository(deps.imageRepo),
		challengecmd.ImageBuildConfig{
			Registry:         registry.Server,
			BuildTimeout:     registry.BuildTimeout,
			BuildConcurrency: registry.BuildConcurrency,
		},
		options...,
	)
}

func buildAWDChallengeHandler(deps moduleDeps, imageBuildService *challengecmd.ImageBuildService) *challengehttp.AWDChallengeHandler {
	importService := challengecmd.NewAWDChallengeImportService(
		deps.awdChallengeCommandRepo,
		challengeinfra.NewAWDChallengeImportPreviewStore(""),
		challengeinfra.NewChallengePackageStorage(challengeinfra.ChallengePackageStorageConfig{}),
		challengeinfra.NewAWDCheckerArtifactStore(""),
		imageBuildService,
	)
	importService.SetTxRunner(NewAWDChallengeImportTxRunner(deps.rawRepo, imageBuildService))
	commandService := challengecmd.NewAWDChallengeCommandFacade(deps.awdChallengeCommandRepo, importService)
	commandService.SetImportLogger(deps.input.Logger.Named("awd_challenge_import_service"))
	queryService := challengeqry.NewAWDChallengeQueryService(deps.awdChallengeQueryRepo)
	return challengehttp.NewAWDChallengeHandler(commandService, queryService)
}

func buildImageHandler(deps moduleDeps) (*challengecmd.ImageService, *challengehttp.ImageHandler) {
	imageCommandRepo := challengeinfra.NewImageCommandRepository(deps.imageRepo)
	imageCommandService := challengecmd.NewImageService(
		imageCommandRepo,
		deps.imageUsageRepo,
		deps.imageRuntime,
		deps.input.Logger.Named("image_service"),
	)
	imageCommandService.StartBackgroundTasks(deps.input.AppContext)
	imageQueryService := challengeqry.NewImageService(challengeinfra.NewImageQueryRepository(deps.imageRepo), deps.input.Config)
	return imageCommandService, challengehttp.NewImageHandler(imageCommandService, imageQueryService)
}

func buildCoreHandler(deps moduleDeps, imageBuildService *challengecmd.ImageBuildService) (*challengepublishcheck.ChallengePublishCheckService, *challengehttp.Handler, interface {
	GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*challengecontracts.ChallengeDetailResp, error)
}) {
	cfg := deps.input.Config
	challengeCommandRepo := challengeinfra.NewChallengeCommandRepository(deps.challengeCommandRepo)
	challengeCommandImageRepo := challengeinfra.NewImageQueryRepository(deps.imageRepo)
	packageRepo := challengeinfra.NewTopologyPackageRevisionRepository(deps.challengeCommandRepo)
	packageStorage := challengeinfra.NewChallengePackageStorage(challengeinfra.ChallengePackageStorageConfig{})
	challengeCommandService := challengecore.NewChallengeService(
		challengeCommandRepo,
		challengeCommandImageRepo,
		deps.topologyRepo,
		deps.input.Logger.Named("challenge_command_service"),
	)
	challengeCommandService.SetEventBus(deps.input.Events)
	challengeImportService := challengeimport.NewChallengeImportService(
		challengeinfra.NewChallengeImportPreviewStore(""),
		challengeinfra.NewChallengeAttachmentStore(
			platformsharedfs.NewStore(deps.input.Config.SharedStoragePath("challenge-attachments")),
			"",
		),
		packageStorage,
		NewChallengeImportTxRunner(deps.rawRepo, imageBuildService),
		imageBuildService,
		deps.input.Events,
		deps.input.Logger.Named("challenge_import_service"),
	)
	challengeSelfCheckService := challengeselfcheck.NewChallengeSelfCheckService(
		challengeCommandRepo,
		challengeCommandImageRepo,
		deps.topologyRepo,
		deps.runtimeProbe,
		challengeselfcheck.Config{
			RuntimeCreateTimeout: cfg.Container.CreateTimeout,
			FlagGlobalSecret:     cfg.Container.FlagGlobalSecret,
		},
		deps.input.Logger.Named("challenge_self_check_service"),
	)
	challengePackageExportService := challengepackageexport.NewChallengePackageExportService(
		challengeCommandRepo,
		deps.topologyRepo,
		packageRepo,
		NewChallengePackageExportTxRunner(deps.rawRepo),
		packageStorage,
	)
	challengePublishCheckService := challengepublishcheck.NewChallengePublishCheckService(
		challengeCommandRepo,
		challengeCommandRepo,
		challengeSelfCheckService,
		challengepublishcheck.Config{
			PollInterval: cfg.Challenge.PublishCheck.PollInterval,
			BatchSize:    cfg.Challenge.PublishCheck.BatchSize,
		},
		deps.input.Events,
		deps.input.Logger.Named("challenge_publish_check_service"),
	)
	challengeQueryService := challengeqry.NewChallengeService(deps.challengeQueryRepo, challengeinfra.NewSolvedCountCache(deps.input.Cache), &challengeqry.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, deps.input.Logger.Named("challenge_service"))
	packageDeliveryService := challengecmd.NewPackageDeliveryService(challengeImportService, nil)
	return challengePublishCheckService, challengehttp.NewHandler(challengehttp.HandlerDeps{
		Commands:       challengeCommandService,
		Queries:        challengeQueryService,
		Imports:        challengeImportService,
		SelfChecks:     challengeSelfCheckService,
		PublishChecks:  challengePublishCheckService,
		PackageExports: challengePackageExportService,
		Attachments: challengeinfra.NewChallengeAttachmentStore(
			platformsharedfs.NewStore(deps.input.Config.SharedStoragePath("challenge-attachments")),
			"",
		),
		PackageDelivery: packageDeliveryService,
	}), challengeQueryService
}

func buildFlagHandler(deps moduleDeps) (*challengehttp.FlagHandler, challengecontracts.FlagValidator, error) {
	flagCommandService, err := challengecmd.NewFlagService(deps.flagRepo, deps.input.Config.Container.FlagGlobalSecret)
	if err != nil {
		return nil, nil, err
	}
	flagQueryService, err := challengeqry.NewFlagService(deps.flagRepo, deps.input.Config.Container.FlagGlobalSecret)
	if err != nil {
		return nil, nil, err
	}
	return challengehttp.NewFlagHandler(flagCommandService, flagQueryService), flagQueryService, nil
}

func buildTopologyHandler(deps moduleDeps) *challengehttp.TopologyHandler {
	topologyRepo := deps.topologyRepo
	templateRepo := challengeinfra.NewTopologyTemplateRepository(deps.templateRepo)
	imageRepo := challengeinfra.NewImageQueryRepository(deps.imageRepo)
	topologyCommandService := challengecmd.NewTopologyService(topologyRepo, templateRepo, imageRepo)
	topologyQueryService := challengeqry.NewTopologyService(topologyRepo, templateRepo)
	return challengehttp.NewTopologyHandler(topologyCommandService, topologyQueryService)
}

func buildWriteupHandler(deps moduleDeps) *challengehttp.WriteupHandler {
	writeupCommandService := challengecmd.NewWriteupService(deps.writeupRepo)
	writeupQueryService := challengeqry.NewWriteupService(deps.writeupRepo)
	return challengehttp.NewWriteupHandler(writeupCommandService, writeupQueryService)
}
