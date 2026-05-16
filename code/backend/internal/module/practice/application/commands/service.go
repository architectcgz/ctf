package commands

import (
	"context"
	"sync"
	"time"

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
	practiceports.PracticeDesiredAWDContestRepository
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
	repo             practiceCommandRepository
	challengeRepo    challengecontracts.PracticeChallengeContract
	contestScope     practiceports.PracticeContestScopeRepository
	imageRepo        challengecontracts.ImageStore
	instanceRepo     instanceRepository
	manualReviewRepo practiceports.PracticeManualReviewRepository
	solvedSubmission practiceports.PracticeSolvedSubmissionRepository
	readinessProbe   practiceports.PracticeInstanceReadinessProbe
	runtimeSubject   practiceports.PracticeRuntimeSubjectRepository
	runtimeService   practiceports.RuntimeInstanceService
	scoreService     ScoreUpdater
	rateLimitStore   practiceports.PracticeFlagSubmitRateLimitStore
	config           *config.Config
	logger           *zap.Logger
	eventBus         platformevents.Bus
	baseCtx          context.Context
	cancel           context.CancelFunc
	tasks            sync.WaitGroup
}

func (s *Service) SetEventBus(bus platformevents.Bus) *Service {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *Service) SetInstanceReadinessProbe(probe practiceports.PracticeInstanceReadinessProbe) *Service {
	if s == nil {
		return nil
	}
	s.readinessProbe = probe
	return s
}

func (s *Service) SetContestScopeRepository(repo practiceports.PracticeContestScopeRepository) *Service {
	if s == nil {
		return nil
	}
	s.contestScope = repo
	return s
}

func (s *Service) SetRuntimeSubjectRepository(repo practiceports.PracticeRuntimeSubjectRepository) *Service {
	if s == nil {
		return nil
	}
	s.runtimeSubject = repo
	return s
}

func (s *Service) SetManualReviewRepository(repo practiceports.PracticeManualReviewRepository) *Service {
	if s == nil {
		return nil
	}
	s.manualReviewRepo = repo
	return s
}

func (s *Service) SetSolvedSubmissionRepository(repo practiceports.PracticeSolvedSubmissionRepository) *Service {
	if s == nil {
		return nil
	}
	s.solvedSubmission = repo
	return s
}

func NewService(
	repo practiceCommandRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
	imageRepo challengecontracts.ImageStore,
	instanceRepo instanceRepository,
	runtimeService practiceports.RuntimeInstanceService,
	scoreService ScoreUpdater,
	rateLimitStore practiceports.PracticeFlagSubmitRateLimitStore,
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
		rateLimitStore: rateLimitStore,
		config:         cfg,
		logger:         logger,
	}
}
