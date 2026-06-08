package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/shared/lockkeepalive"
)

type redisLockKeepaliveConfig struct {
	Name string
	TTL  time.Duration
}

func startRedisLockKeepalive(ctx context.Context, log *zap.Logger, lock contestports.ContestSchedulerLockLease, cfg redisLockKeepaliveConfig) (context.Context, func()) {
	return lockkeepalive.Start(ctx, log, lock, cfg.Name, cfg.TTL)
}

func redisLockRefreshInterval(ttl time.Duration) time.Duration {
	return lockkeepalive.RefreshInterval(ttl)
}
