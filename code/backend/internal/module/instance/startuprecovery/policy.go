package startuprecovery

import (
	"time"

	"ctf-platform/internal/shared/lockkeepalive"
)

const (
	DefaultHeartbeatInterval = 30 * time.Second
	DefaultLockTTL           = 40 * time.Second
	DefaultLeaderRetry       = time.Second

	heartbeatToleranceFactor = 2
)

func NormalizeHeartbeatInterval(interval time.Duration) time.Duration {
	if interval <= 0 {
		return DefaultHeartbeatInterval
	}
	return interval
}

func NormalizeLeaderRetry(interval time.Duration) time.Duration {
	if interval <= 0 {
		return DefaultLeaderRetry
	}
	return interval
}

func HeartbeatStaleThreshold(interval time.Duration) time.Duration {
	return NormalizeHeartbeatInterval(interval) * heartbeatToleranceFactor
}

func LockFailoverWindow(ttl, retryInterval time.Duration) time.Duration {
	return lockkeepalive.FailoverWindow(ttl, NormalizeLeaderRetry(retryInterval))
}

func MaxSafeLockTTL(heartbeatInterval, retryInterval time.Duration) time.Duration {
	threshold := HeartbeatStaleThreshold(heartbeatInterval)
	retry := NormalizeLeaderRetry(retryInterval)
	if threshold <= retry {
		return 0
	}

	lo := time.Duration(0)
	hi := threshold - retry
	for lo < hi {
		mid := lo + (hi-lo+1)/2
		if LockFailoverWindow(mid, retry) <= threshold {
			lo = mid
			continue
		}
		hi = mid - 1
	}
	// Config values are operated at coarse granularity. Round the exact
	// nanosecond-safe upper bound down to a whole second so the exported limit
	// stays stable and human-readable.
	if lo >= time.Second {
		return lo.Truncate(time.Second)
	}
	return lo
}
