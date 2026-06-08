package lockkeepalive

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Lease interface {
	Refresh(ctx context.Context, ttl time.Duration) (bool, error)
	Key(ctx context.Context) string
}

// RefreshInterval returns how often a live owner should renew a lease with the
// given TTL. The current policy keeps at least two chances to refresh before
// the lease expires, while avoiding zero or negative durations for tiny TTLs.
func RefreshInterval(ttl time.Duration) time.Duration {
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

// FailoverWindow returns the worst-case delay between the last confirmed owner
// heartbeat and the next leader taking over the same lease:
//  1. wait for the old lease TTL to expire,
//  2. plus one keepalive refresh interval because the old leader may die
//     immediately after a successful refresh but before the next heartbeat,
//  3. plus one leader-election retry tick before a standby retries acquisition.
func FailoverWindow(ttl, retryInterval time.Duration) time.Duration {
	if ttl <= 0 {
		return 0
	}
	if retryInterval <= 0 {
		retryInterval = time.Millisecond
	}
	return ttl + RefreshInterval(ttl) + retryInterval
}

func Start(ctx context.Context, log *zap.Logger, lease Lease, name string, ttl time.Duration) (context.Context, func()) {
	if log == nil {
		log = zap.NewNop()
	}
	if lease == nil || ttl <= 0 {
		return ctx, func() {}
	}

	runCtx, runCancel := context.WithCancel(ctx)
	keepaliveCtx, keepaliveCancel := context.WithCancel(runCtx)
	done := make(chan struct{})
	interval := RefreshInterval(ttl)
	lastConfirmedAt := time.Now()

	go func() {
		ticker := time.NewTicker(interval)
		defer close(done)
		defer ticker.Stop()

		for {
			select {
			case <-keepaliveCtx.Done():
				return
			case <-ticker.C:
				remaining := time.Until(lastConfirmedAt.Add(ttl))
				if remaining <= 0 {
					log.Warn("distributed_lock_refresh_deadline_exceeded",
						zap.String("lock_name", name),
						zap.String("lock_key", lease.Key(keepaliveCtx)),
						zap.Duration("ttl", ttl))
					runCancel()
					return
				}

				refreshCtx, cancel := context.WithTimeout(keepaliveCtx, refreshTimeout(interval, remaining))
				refreshed, err := lease.Refresh(refreshCtx, ttl)
				cancel()
				if err != nil {
					if keepaliveCtx.Err() == nil {
						log.Error("distributed_lock_refresh_failed",
							zap.String("lock_name", name),
							zap.String("lock_key", lease.Key(refreshCtx)),
							zap.Error(err))
					}
					if keepaliveCtx.Err() == nil && time.Since(lastConfirmedAt) >= ttl {
						log.Warn("distributed_lock_refresh_deadline_exceeded",
							zap.String("lock_name", name),
							zap.String("lock_key", lease.Key(refreshCtx)),
							zap.Duration("ttl", ttl))
						runCancel()
						return
					}
					continue
				}
				if refreshed {
					lastConfirmedAt = time.Now()
					continue
				}

				log.Warn("distributed_lock_lost_during_run",
					zap.String("lock_name", name),
					zap.String("lock_key", lease.Key(keepaliveCtx)))
				runCancel()
				return
			}
		}
	}()

	return runCtx, func() {
		keepaliveCancel()
		<-done
	}
}

func refreshTimeout(interval, remaining time.Duration) time.Duration {
	timeout := interval
	if timeout <= 0 || remaining < timeout {
		timeout = remaining
	}
	if timeout <= 0 {
		timeout = time.Millisecond
	}
	return timeout
}
