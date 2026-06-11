package commands

import (
	"context"
	"time"

	challengecore "ctf-platform/internal/module/challenge/application/challengecore"
	challengeimport "ctf-platform/internal/module/challenge/application/challengeimport"
	challengepackageexport "ctf-platform/internal/module/challenge/application/challengepackageexport"
	challengepublishcheck "ctf-platform/internal/module/challenge/application/challengepublishcheck"
	challengeselfcheck "ctf-platform/internal/module/challenge/application/challengeselfcheck"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"

	"go.uber.org/zap"
)

type ChallengeService = challengecore.ChallengeService
type ChallengeHintInput = challengecore.ChallengeHintInput
type CreateChallengeInput = challengecore.CreateChallengeInput
type UpdateChallengeInput = challengecore.UpdateChallengeInput

type ChallengeImportService = challengeimport.ChallengeImportService
type ChallengeSelfCheckService = challengeselfcheck.ChallengeSelfCheckService
type ChallengePublishCheckService = challengepublishcheck.ChallengePublishCheckService
type ChallengePackageExportService = challengepackageexport.ChallengePackageExportService

type SelfCheckConfig struct {
	RuntimeCreateTimeout     time.Duration
	FlagGlobalSecret         string
	PublishCheckPollInterval time.Duration
	PublishCheckBatchSize    int
}

type splitTestChallengeCommandRepository interface {
	challengeports.ChallengeWriteRepository
	challengeports.ChallengeInstanceUsageRepository
}

type splitTestChallengeLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*challengeports.ChallengeWriteModel, error)
}

func NewChallengeService(
	repo splitTestChallengeCommandRepository,
	imageRepo challengeports.ImageQueryRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	logger *zap.Logger,
) *ChallengeService {
	return challengecore.NewChallengeService(repo, imageRepo, topologyRepo, logger)
}

func NewChallengeImportService(
	previewStore challengeports.ChallengeImportPreviewStore,
	attachmentStore challengeports.ChallengeAttachmentStore,
	packageStorage challengeports.ChallengePackageStorage,
	txRunner challengeports.ChallengeImportTxRunner,
	imageBuild challengeimport.ImageBuildService,
	eventBus platformevents.Bus,
	logger *zap.Logger,
) *ChallengeImportService {
	return challengeimport.NewChallengeImportService(previewStore, attachmentStore, packageStorage, txRunner, imageBuild, eventBus, logger)
}

func NewChallengeSelfCheckService(
	repo splitTestChallengeLookupRepository,
	imageRepo challengeports.ImageQueryRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
	cfg SelfCheckConfig,
	logger *zap.Logger,
) *ChallengeSelfCheckService {
	return challengeselfcheck.NewChallengeSelfCheckService(repo, imageRepo, topologyRepo, runtimeProbe, challengeselfcheck.Config{
		RuntimeCreateTimeout: cfg.RuntimeCreateTimeout,
		FlagGlobalSecret:     cfg.FlagGlobalSecret,
	}, logger)
}

func NewChallengePublishCheckService(
	challengeRepo splitTestChallengeLookupRepository,
	jobRepo challengeports.ChallengePublishCheckRepository,
	selfChecker challengepublishcheck.ChallengeSelfChecker,
	publisher challengepublishcheck.ChallengePublisher,
	cfg SelfCheckConfig,
	eventBus platformevents.Bus,
	logger *zap.Logger,
) *ChallengePublishCheckService {
	return challengepublishcheck.NewChallengePublishCheckService(challengeRepo, jobRepo, selfChecker, publisher, challengepublishcheck.Config{
		PollInterval: cfg.PublishCheckPollInterval,
		BatchSize:    cfg.PublishCheckBatchSize,
	}, eventBus, logger)
}

func NewChallengePackageExportService(
	challengeRepo splitTestChallengeLookupRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	packageRepo challengeports.ChallengePackageRevisionRepository,
	txRunner challengeports.ChallengePackageExportTxRunner,
	packageStorage challengeports.ChallengePackageStorage,
) *ChallengePackageExportService {
	return challengepackageexport.NewChallengePackageExportService(challengeRepo, topologyRepo, packageRepo, txRunner, packageStorage)
}

func stringPointer(value string) *string {
	return &value
}
