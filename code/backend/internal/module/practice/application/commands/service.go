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

type practiceSchedulerLockLease interface {
	Key(ctx context.Context) string
	Release(ctx context.Context) (bool, error)
	Refresh(ctx context.Context, ttl time.Duration) (bool, error)
}

type practiceCommandRepository interface {
	practiceports.PracticeInstanceStartTxManager
	practiceports.PracticeInstanceRestartTxManager
	practiceports.PracticeAWDServiceOperationTxManager
	practiceports.PracticeSubmissionOutboxTxManager
	practiceports.PracticeInstanceStartTxRepository
	practiceports.PracticeInstanceRestartTxRepository
	practiceports.PracticeAWDServiceOperationTxRepository
	practiceports.PracticeContestLookupRepository
	practiceports.PracticeDesiredAWDContestRepository
	practiceports.PracticeContestChallengeLookupRepository
	practiceports.PracticeContestAWDServiceRepository
	practiceports.PracticeContestAWDServiceRuntimeSubjectRepository
	practiceports.PracticeContestAWDInstanceRepository
	practiceports.PracticeContestTeamRepository
	practiceports.PracticeContestRegistrationRepository
	practiceports.PracticeAWDScopeControlRepository
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

type serviceCore struct {
	repo                practiceCommandRepository
	contestScope        practiceports.PracticeContestScopeRepository
	imageRepo           challengecontracts.ImageStore
	instanceRepo        instanceRepository
	manualReviewRepo    practiceports.PracticeManualReviewRepository
	solvedSubmission    practiceports.PracticeSolvedSubmissionRepository
	readinessProbe      practiceports.PracticeInstanceReadinessProbe
	runtimeSubject      practiceports.PracticeRuntimeSubjectRepository
	runtimeNodeSelector practiceports.RuntimeNodeSelector
	runtimeService      practiceports.RuntimeInstanceService
	scoreService        ScoreUpdater
	rateLimitStore      practiceports.PracticeFlagSubmitRateLimitStore
	desiredState        practiceports.PracticeDesiredAWDReconcileStateStore
	schedulerLockStore  practiceports.PracticeInstanceSchedulerLockStore
	config              *config.Config
	logger              *zap.Logger
	eventBus            platformevents.Bus
	baseCtx             context.Context
	cancel              context.CancelFunc
	tasks               sync.WaitGroup
}

func (s *serviceCore) SetEventBus(bus platformevents.Bus) *serviceCore {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *serviceCore) SetDesiredAWDReconcileStateStore(store practiceports.PracticeDesiredAWDReconcileStateStore) *serviceCore {
	if s == nil {
		return nil
	}
	s.desiredState = store
	return s
}

func (s *serviceCore) SetSchedulerLockStore(store practiceports.PracticeInstanceSchedulerLockStore) *serviceCore {
	if s == nil {
		return nil
	}
	s.schedulerLockStore = store
	return s
}

func (s *serviceCore) SetInstanceReadinessProbe(probe practiceports.PracticeInstanceReadinessProbe) *serviceCore {
	if s == nil {
		return nil
	}
	s.readinessProbe = probe
	return s
}

func (s *serviceCore) SetContestScopeRepository(repo practiceports.PracticeContestScopeRepository) *serviceCore {
	if s == nil {
		return nil
	}
	s.contestScope = repo
	return s
}

func (s *serviceCore) SetRuntimeSubjectRepository(repo practiceports.PracticeRuntimeSubjectRepository) *serviceCore {
	if s == nil {
		return nil
	}
	s.runtimeSubject = repo
	return s
}

func (s *serviceCore) SetRuntimeNodeSelector(selector practiceports.RuntimeNodeSelector) *serviceCore {
	if s == nil {
		return nil
	}
	s.runtimeNodeSelector = selector
	return s
}

func (s *serviceCore) SetManualReviewRepository(repo practiceports.PracticeManualReviewRepository) *serviceCore {
	if s == nil {
		return nil
	}
	s.manualReviewRepo = repo
	return s
}

func (s *serviceCore) SetSolvedSubmissionRepository(repo practiceports.PracticeSolvedSubmissionRepository) *serviceCore {
	if s == nil {
		return nil
	}
	s.solvedSubmission = repo
	return s
}

func newServiceCore(
	repo practiceCommandRepository,
	imageRepo challengecontracts.ImageStore,
	instanceRepo instanceRepository,
	runtimeService practiceports.RuntimeInstanceService,
	scoreService ScoreUpdater,
	rateLimitStore practiceports.PracticeFlagSubmitRateLimitStore,
	cfg *config.Config,
	logger *zap.Logger,
) *serviceCore {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &serviceCore{
		repo:           repo,
		imageRepo:      imageRepo,
		instanceRepo:   instanceRepo,
		runtimeService: runtimeService,
		scoreService:   scoreService,
		rateLimitStore: rateLimitStore,
		config:         cfg,
		logger:         logger,
	}
}
