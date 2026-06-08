package startuprecovery

import (
	"testing"
	"time"
)

func TestHeartbeatStaleThresholdUsesSharedDefault(t *testing.T) {
	t.Parallel()

	if got := HeartbeatStaleThreshold(0); got != time.Minute {
		t.Fatalf("expected default stale threshold 1m0s, got %s", got)
	}
}

func TestMaxSafeLockTTLMatchesDefaultPolicy(t *testing.T) {
	t.Parallel()

	if got := MaxSafeLockTTL(DefaultHeartbeatInterval, DefaultLeaderRetry); got != 44*time.Second {
		t.Fatalf("expected max safe lock ttl 44s, got %s", got)
	}
}
