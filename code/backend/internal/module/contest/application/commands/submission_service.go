package commands

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

type scoreboardUpdater interface {
	UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error
	RebuildScoreboard(ctx context.Context, contestID int64) error
}

type SubmissionService struct {
	contestRepo       contestports.Repository
	repo              contestports.ContestSubmissionRepository
	redis             *redislib.Client
	flagValidator     challengecontracts.FlagValidator
	teamRepo          contestports.ContestTeamFinder
	scoreboardService scoreboardUpdater
	cfg               *config.Config
}

func NewSubmissionService(contestRepo contestports.Repository, repo contestports.ContestSubmissionRepository, redis *redislib.Client, flagValidator challengecontracts.FlagValidator, teamRepo contestports.ContestTeamFinder, scoreboardService scoreboardUpdater, cfg *config.Config) *SubmissionService {
	return &SubmissionService{
		contestRepo:       contestRepo,
		repo:              repo,
		redis:             redis,
		flagValidator:     flagValidator,
		teamRepo:          teamRepo,
		scoreboardService: scoreboardService,
		cfg:               cfg,
	}
}
