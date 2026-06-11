package commands

import (
	"context"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
	"go.uber.org/zap"
)

type scoreboardUpdater interface {
	UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error
	RebuildScoreboard(ctx context.Context, contestID int64) error
}

type submissionRepository interface {
	contestports.ContestSubmissionScoringTxRunner
	contestports.ContestSubmissionRegistrationLookupRepository
	contestports.ContestSubmissionChallengeLookupRepository
	contestports.ContestSubmissionWriteRepository
}

type SubmissionService struct {
	contestRepo       contestports.ContestLookupRepository
	repo              submissionRepository
	rateLimitStore    contestports.ContestSubmissionRateLimitStore
	flagValidator     challengecontracts.FlagValidator
	teamRepo          contestports.ContestTeamFinder
	scoreboardService scoreboardUpdater
	eventBus          platformevents.Bus
	cfg               *config.Config
	log               *zap.Logger
}

func NewSubmissionService(contestRepo contestports.ContestLookupRepository, repo submissionRepository, rateLimitStore contestports.ContestSubmissionRateLimitStore, flagValidator challengecontracts.FlagValidator, teamRepo contestports.ContestTeamFinder, scoreboardService scoreboardUpdater, cfg *config.Config, logger *zap.Logger) *SubmissionService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SubmissionService{
		contestRepo:       contestRepo,
		repo:              repo,
		rateLimitStore:    rateLimitStore,
		flagValidator:     flagValidator,
		teamRepo:          teamRepo,
		scoreboardService: scoreboardService,
		cfg:               cfg,
		log:               logger,
	}
}

func (s *SubmissionService) SetEventBus(bus platformevents.Bus) *SubmissionService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *SubmissionService) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
	if s == nil || s.eventBus == nil {
		return
	}
	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.log.Warn("publish_contest_event_failed", zap.String("event", evt.Name), zap.Error(err))
	}
}
