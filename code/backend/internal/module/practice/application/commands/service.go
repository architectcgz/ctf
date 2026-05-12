package commands

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
)

const errMsgChallengeNoTarget = "该题目不需要靶机实例"

type ScoreUpdater interface {
	UpdateUserScore(ctx context.Context, userID int64) error
	lockTimeout() time.Duration
}

type practiceCommandRepository interface {
	practiceports.PracticeInstanceStartTxManager
	practiceports.PracticeInstanceRestartTxManager
	practiceports.PracticeAWDServiceOperationTxManager
	practiceports.PracticeInstanceStartTxRepository
	practiceports.PracticeInstanceRestartTxRepository
	practiceports.PracticeAWDServiceOperationTxRepository
	practiceports.PracticeContestLookupRepository
	practiceports.PracticeContestChallengeLookupRepository
	practiceports.PracticeContestAWDServiceRepository
	practiceports.PracticeContestAWDInstanceRepository
	practiceports.PracticeContestTeamRepository
	practiceports.PracticeContestRegistrationRepository
	practiceports.PracticeSubmissionWriteRepository
	practiceports.PracticeSolvedSubmissionRepository
	practiceports.PracticeSubmissionHistoryRepository
	practiceports.PracticeSubmissionConstraintRepository
	practiceports.PracticeUserLookupRepository
	practiceports.PracticeManualReviewListRepository
	practiceports.PracticeManualReviewLookupRepository
}

type instanceRepository interface {
	practiceports.PracticeInstanceLookupRepository
	practiceports.PracticeInstanceRuntimeWriteRepository
	practiceports.PracticeInstanceAWDOperationRepository
	practiceports.PracticeInstanceStatusRepository
	practiceports.PracticePendingInstanceRepository
	practiceports.PracticeInstanceStatsRepository
}

type Service struct {
	repo           practiceCommandRepository
	challengeRepo  challengecontracts.PracticeChallengeContract
	imageRepo      challengecontracts.ImageStore
	instanceRepo   instanceRepository
	runtimeService practiceports.RuntimeInstanceService
	scoreService   ScoreUpdater
	redis          *redis.Client
	config         *config.Config
	logger         *zap.Logger
	eventBus       platformevents.Bus
	baseCtx        context.Context
	cancel         context.CancelFunc
	tasks          sync.WaitGroup
}

func (s *Service) SetEventBus(bus platformevents.Bus) *Service {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func NewService(
	repo practiceCommandRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
	imageRepo challengecontracts.ImageStore,
	instanceRepo instanceRepository,
	runtimeService practiceports.RuntimeInstanceService,
	scoreService ScoreUpdater,
	redis *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &Service{
		repo:           repo,
		challengeRepo:  challengeRepo,
		imageRepo:      imageRepo,
		instanceRepo:   instanceRepo,
		runtimeService: runtimeService,
		scoreService:   scoreService,
		redis:          redis,
		config:         cfg,
		logger:         logger,
	}
}
