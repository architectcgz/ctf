package jobs

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
)

type StatusUpdater struct {
	repo      contestports.ContestStatusRepository
	awdRepo   contestports.AWDRepository
	redis     *redislib.Client
	log       *zap.Logger
	interval  time.Duration
	batchSize int
	lockTTL   time.Duration
}

func NewStatusUpdater(repo contestports.ContestStatusRepository, redis *redislib.Client, interval time.Duration, batchSize int, lockTTL time.Duration, log *zap.Logger, awdRepos ...contestports.AWDRepository) *StatusUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	if lockTTL <= 0 {
		lockTTL = 30 * time.Second
	}
	var awdRepo contestports.AWDRepository
	if len(awdRepos) > 0 {
		awdRepo = awdRepos[0]
	}
	return &StatusUpdater{
		repo:      repo,
		awdRepo:   awdRepo,
		redis:     redis,
		log:       log,
		interval:  interval,
		batchSize: batchSize,
		lockTTL:   lockTTL,
	}
}

func (u *StatusUpdater) Start(ctx context.Context) {
	u.updateStatuses(ctx)

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.updateStatuses(ctx)
		}
	}
}
