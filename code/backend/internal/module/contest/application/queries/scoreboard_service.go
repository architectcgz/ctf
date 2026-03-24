package queries

import (
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ScoreboardService struct {
	repo   contestports.Repository
	redis  *redislib.Client
	logger *zap.Logger
	config *config.ContestConfig
}

func NewScoreboardService(repo contestports.Repository, redis *redislib.Client, cfg *config.ContestConfig, logger *zap.Logger) *ScoreboardService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ScoreboardService{
		repo:   repo,
		redis:  redis,
		logger: logger,
		config: cfg,
	}
}
