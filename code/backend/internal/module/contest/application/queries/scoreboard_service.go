package queries

import (
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ScoreboardService struct {
	repo       contestports.ContestScoreboardRepository
	stateStore contestports.ContestScoreboardStateStore
	logger     *zap.Logger
	config     *config.ContestConfig
}

func NewScoreboardService(repo contestports.ContestScoreboardRepository, stateStore contestports.ContestScoreboardStateStore, cfg *config.ContestConfig, logger *zap.Logger) *ScoreboardService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ScoreboardService{
		repo:       repo,
		stateStore: stateStore,
		logger:     logger,
		config:     cfg,
	}
}
