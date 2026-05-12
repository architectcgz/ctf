package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
)

type redisLockKeepaliveConfig struct {
	Name string
	TTL  time.Duration
}

func startRedisLockKeepalive(ctx context.Context, log *zap.Logger, lock contestports.ContestStatusUpdateLockLease, cfg redisLockKeepaliveConfig) (context.Context, func()) {
	if log == nil {
		log = zap.NewNop()
	}
	if lock == nil || cfg.TTL <= 0 {
		return ctx, func() {}
	}

	runCtx, runCancel := context.WithCancel(ctx)
	keepaliveCtx, keepaliveCancel := context.WithCancel(runCtx)
	done := make(chan struct{})
	interval := redisLockRefreshInterval(cfg.TTL)

	go func() {
		ticker := time.NewTicker(interval)
		defer close(done)
		defer ticker.Stop()

		for {
			select {
			case <-keepaliveCtx.Done():
				return
			case <-ticker.C:
				// Long-running schedulers must renew their lease so other instances do not observe an expired lock mid-run.
				refreshed, err := lock.Refresh(keepaliveCtx, cfg.TTL)
				if err != nil {
					if keepaliveCtx.Err() == nil {
						log.Error("scheduler_lock_refresh_failed", zap.String("lock_name", cfg.Name), zap.String("lock_key", lock.Key()), zap.Error(err))
					}
					continue
				}
				if refreshed {
					continue
				}
				// Losing the token means another instance may continue the same job, so this run must stop making progress.
				log.Warn("scheduler_lock_lost_during_run", zap.String("lock_name", cfg.Name), zap.String("lock_key", lock.Key()))
				runCancel()
				return
			}
		}
	}()

	return runCtx, func() {
		keepaliveCancel()
		// Wait for the goroutine to stop so deferred release runs after the last refresh attempt has finished.
		<-done
	}
}

func redisLockRefreshInterval(ttl time.Duration) time.Duration {
	var interval time.Duration
	switch {
	case ttl <= 3*time.Second:
		interval = ttl / 2
	default:
		interval = ttl / 3
	}
	if interval <= 0 {
		interval = ttl
	}
	if interval <= 0 {
		interval = time.Millisecond
	}
	return interval
}
