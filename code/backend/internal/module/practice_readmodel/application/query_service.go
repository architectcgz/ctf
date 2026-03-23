package application

import (
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	practicereadmodel "ctf-platform/internal/module/practice_readmodel"
)

type QueryService struct {
	repo     QueryRepository
	cache    *redis.Client
	cacheTTL time.Duration
	logger   *zap.Logger
}

var _ practicereadmodel.PracticeQuery = (*QueryService)(nil)

func NewQueryService(repo QueryRepository, cache *redis.Client, cacheTTL time.Duration, logger *zap.Logger) *QueryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QueryService{
		repo:     repo,
		cache:    cache,
		cacheTTL: cacheTTL,
		logger:   logger,
	}
}
