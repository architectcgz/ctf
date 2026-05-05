package composition

import (
	"context"
	"fmt"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	"gorm.io/gorm"
)

type BackgroundTaskCloser interface {
	Close(ctx context.Context) error
}

type ChallengeModule struct {
	BackgroundTasks     BackgroundTaskCloser
	AWDChallengeHandler *challengehttp.AWDChallengeHandler
	AWDChallengeQuery   challengeports.AWDChallengeQueryRepository
	Catalog             challengecontracts.ChallengeContract
	FlagHandler         *challengehttp.FlagHandler
	FlagValidator       challengecontracts.FlagValidator
	Handler             *challengehttp.Handler
	ImageHandler        *challengehttp.ImageHandler
	ImageStore          challengecontracts.ImageStore
	TopologyHandler     *challengehttp.TopologyHandler
	WriteupHandler      *challengehttp.WriteupHandler
}

type challengeModuleDeps struct {
	catalog                 challengecontracts.ChallengeContract
	imageStore              challengecontracts.ImageStore
	imageRepo               *challengeinfra.ImageRepository
	challengeCommandRepo    *challengeinfra.Repository
	challengeQueryRepo      *challengeinfra.Repository
	awdChallengeCommandRepo challengeports.AWDChallengeCommandRepository
	awdChallengeQueryRepo   challengeports.AWDChallengeQueryRepository
	flagRepo                challengeports.ChallengeFlagRepository
	imageUsageRepo          challengeports.ChallengeImageUsageRepository
	topologyRepo            *challengeinfra.Repository
	writeupRepo             *challengeinfra.Repository
	templateRepo            *challengeinfra.TemplateRepository
	imageRuntime            challengeports.ImageRuntime
	runtimeProbe            challengeports.ChallengeRuntimeProbe
}

func BuildChallengeModule(root *Root, runtime *RuntimeModule, ops *OpsModule) (*ChallengeModule, error) {
	cfg := root.Config()
	deps := buildChallengeModuleDeps(root, runtime)

	imageCommandService, imageHandler := buildChallengeImageHandler(root, deps)
	imageBuildService := buildChallengeImageBuildService(root, deps)
	coreService, coreHandler := buildChallengeCoreHandler(root, deps, ops, imageBuildService)
	flagHandler, flagValidator, err := buildChallengeFlagHandler(cfg, deps)
	if err != nil {
		return nil, err
	}
	awdChallengeHandler := buildChallengeAWDChallengeHandler(root.DB(), deps, imageBuildService)
	topologyHandler := buildChallengeTopologyHandler(deps)
	writeupHandler := buildChallengeWriteupHandler(deps)
	if root.Config().Challenge.PublishCheck.Enabled {
		root.RegisterBackgroundJob(NewLoopBackgroundJob("challenge_publish_check_worker", coreService.RunPublishCheckLoop))
	}
	if cfg.Container.Registry.BuildEnabled {
		root.RegisterBackgroundJob(NewLoopBackgroundJob("image_build_worker", imageBuildService.RunBuildLoop))
	}

	return &ChallengeModule{
		BackgroundTasks:     imageCommandService,
		AWDChallengeHandler: awdChallengeHandler,
		AWDChallengeQuery:   deps.awdChallengeQueryRepo,
		Catalog:             deps.catalog,
		FlagHandler:         flagHandler,
		FlagValidator:       flagValidator,
		Handler:             coreHandler,
		ImageHandler:        imageHandler,
		ImageStore:          deps.imageStore,
		TopologyHandler:     topologyHandler,
		WriteupHandler:      writeupHandler,
	}, nil
}

func buildChallengeAWDChallengeHandler(db *gorm.DB, deps challengeModuleDeps, imageBuildService *challengecmd.ImageBuildService) *challengehttp.AWDChallengeHandler {
	commandService := challengecmd.NewAWDChallengeCommandFacade(db, deps.awdChallengeCommandRepo, imageBuildService)
	queryService := challengeqry.NewAWDChallengeQueryService(deps.awdChallengeQueryRepo)
	return challengehttp.NewAWDChallengeHandler(commandService, queryService)
}

func buildChallengeModuleDeps(root *Root, runtime *RuntimeModule) challengeModuleDeps {
	db := root.DB()

	challengeRepo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)

	return challengeModuleDeps{
		catalog:                 challengeRepo,
		imageStore:              imageRepo,
		imageRepo:               imageRepo,
		challengeCommandRepo:    challengeRepo,
		challengeQueryRepo:      challengeRepo,
		awdChallengeCommandRepo: challengeRepo,
		awdChallengeQueryRepo:   challengeRepo,
		flagRepo:                challengeRepo,
		imageUsageRepo:          challengeRepo,
		topologyRepo:            challengeRepo,
		writeupRepo:             challengeRepo,
		templateRepo:            challengeinfra.NewTemplateRepository(db),
		imageRuntime:            runtime.ChallengeImageRuntime,
		runtimeProbe:            runtime.ChallengeRuntimeProbe,
	}
}

func buildChallengeImageHandler(root *Root, deps challengeModuleDeps) (*challengecmd.ImageService, *challengehttp.ImageHandler) {
	imageCommandService := challengecmd.NewImageService(
		deps.imageRepo,
		deps.imageUsageRepo,
		deps.imageRuntime,
		root.Logger().Named("image_service"),
	)
	imageCommandService.StartBackgroundTasks(root.Context())
	imageQueryService := challengeqry.NewImageService(deps.imageRepo, root.Config())
	return imageCommandService, challengehttp.NewImageHandler(imageCommandService, imageQueryService)
}

func buildChallengeImageBuildService(root *Root, deps challengeModuleDeps) *challengecmd.ImageBuildService {
	registry := root.Config().Container.Registry
	return challengecmd.NewImageBuildService(
		deps.imageRepo,
		challengecmd.ImageBuildConfig{
			Registry:         registry.Server,
			BuildTimeout:     registry.BuildTimeout,
			BuildConcurrency: registry.BuildConcurrency,
			BatchSize:        registry.BuildConcurrency,
		},
		challengecmd.WithImageBuildDockerBuilder(challengecmd.NewDockerCLIImageBuilderWithConfig(challengecmd.DockerCLIImageBuilderConfig{
			RegistryServer: registry.Server,
			Username:       registry.Username,
			Password:       registry.Password,
			IdentityToken:  registry.IdentityToken,
		})),
		challengecmd.WithImageBuildRegistryVerifier(challengecmd.NewRegistryClient(challengecmd.RegistryClientConfig{
			Scheme:        registry.Scheme,
			Server:        registry.Server,
			Username:      registry.Username,
			Password:      registry.Password,
			IdentityToken: registry.IdentityToken,
		}, nil)),
		challengecmd.WithImageBuildLogger(root.Logger().Named("image_build_service")),
	)
}

type challengePublishNotificationSender struct {
	service *opscmd.NotificationService
}

func (s *challengePublishNotificationSender) SendChallengePublishCheckResult(ctx context.Context, userID int64, challengeID int64, challengeTitle string, passed bool, failureSummary string) error {
	if s == nil || s.service == nil {
		return nil
	}
	title := "题目发布失败"
	content := fmt.Sprintf("《%s》未通过平台自检。", challengeTitle)
	if passed {
		title = "题目发布成功"
		content = fmt.Sprintf("《%s》已通过平台自检并自动发布。", challengeTitle)
	} else if failureSummary != "" {
		content = fmt.Sprintf("《%s》未通过平台自检：%s", challengeTitle, failureSummary)
	}
	link := fmt.Sprintf("/admin/challenges/%d", challengeID)
	return s.service.SendNotification(ctx, userID, &dto.NotificationReq{
		Type:    model.NotificationTypeChallenge,
		Title:   title,
		Content: content,
		Link:    &link,
	})
}

func buildChallengeCoreHandler(root *Root, deps challengeModuleDeps, ops *OpsModule, imageBuildService *challengecmd.ImageBuildService) (*challengecmd.ChallengeService, *challengehttp.Handler) {
	var notifications challengecmd.ChallengeNotificationSender = nil
	if ops != nil {
		notifications = &challengePublishNotificationSender{
			service: opscmd.NewNotificationService(
				opsinfra.NewNotificationRepository(root.DB()),
				root.Config().Pagination,
				ops.WebSocketManager,
				root.Logger().Named("challenge_publish_notification_service"),
			),
		}
	}
	challengeCommandService := challengecmd.NewChallengeService(
		root.DB(),
		deps.challengeCommandRepo,
		deps.imageRepo,
		deps.topologyRepo,
		deps.topologyRepo,
		deps.runtimeProbe,
		challengecmd.SelfCheckConfig{
			RuntimeCreateTimeout:     root.Config().Container.CreateTimeout,
			FlagGlobalSecret:         root.Config().Container.FlagGlobalSecret,
			PublishCheckPollInterval: root.Config().Challenge.PublishCheck.PollInterval,
			PublishCheckBatchSize:    root.Config().Challenge.PublishCheck.BatchSize,
		},
		root.Logger().Named("challenge_command_service"),
		notifications,
	)
	challengeCommandService.SetImageBuildService(imageBuildService)
	challengeQueryService := challengeqry.NewChallengeService(deps.challengeQueryRepo, root.Cache(), &challengeqry.Config{
		SolvedCountCacheTTL: root.Config().Challenge.SolvedCountCacheTTL,
	}, root.Logger().Named("challenge_service"))
	return challengeCommandService, challengehttp.NewHandler(challengeCommandService, challengeQueryService)
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
	topologyQueryService := challengeqry.NewTopologyService(deps.topologyRepo, deps.templateRepo, deps.topologyRepo)
	return challengehttp.NewTopologyHandler(topologyCommandService, topologyQueryService)
}

func buildChallengeWriteupHandler(deps challengeModuleDeps) *challengehttp.WriteupHandler {
	writeupCommandService := challengecmd.NewWriteupService(deps.writeupRepo)
	writeupQueryService := challengeqry.NewWriteupService(deps.writeupRepo)
	return challengehttp.NewWriteupHandler(writeupCommandService, writeupQueryService)
}
