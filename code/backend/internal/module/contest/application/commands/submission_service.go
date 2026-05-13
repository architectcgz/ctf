package commands

import (
	"context"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
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
}

func NewSubmissionService(contestRepo contestports.ContestLookupRepository, repo submissionRepository, rateLimitStore contestports.ContestSubmissionRateLimitStore, flagValidator challengecontracts.FlagValidator, teamRepo contestports.ContestTeamFinder, scoreboardService scoreboardUpdater, cfg *config.Config) *SubmissionService {
	return &SubmissionService{
		contestRepo:       contestRepo,
		repo:              repo,
		rateLimitStore:    rateLimitStore,
		flagValidator:     flagValidator,
		teamRepo:          teamRepo,
		scoreboardService: scoreboardService,
		cfg:               cfg,
	}
}

func (s *SubmissionService) SetEventBus(bus platformevents.Bus) *SubmissionService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}
