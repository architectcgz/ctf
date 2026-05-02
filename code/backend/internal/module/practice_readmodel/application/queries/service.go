package queries

import (
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/module/practice_readmodel/ports"
)

type queryRepository interface {
	ports.ProgressQueryRepository
	ports.TimelineQueryRepository
}

type QueryService struct {
	repo     queryRepository
	cache    *redis.Client
	cacheTTL time.Duration
	logger   *zap.Logger
}

var _ Service = (*QueryService)(nil)

func NewQueryService(repo queryRepository, cache *redis.Client, cacheTTL time.Duration, logger *zap.Logger) *QueryService {
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
