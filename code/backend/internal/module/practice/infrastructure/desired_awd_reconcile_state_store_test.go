package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	practiceports "ctf-platform/internal/module/practice/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func TestDesiredAWDReconcileStateStoreAppliesBoundedTTL(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	store := NewDesiredAWDReconcileStateStore(redisClient)
	now := time.Now().UTC()
	state := &practiceports.DesiredAWDReconcileState{
		FailureCount:  2,
		LastFailureAt: now,
		NextAttemptAt: now.Add(5 * time.Minute),
	}

	if err := store.StoreDesiredAWDReconcileState(context.Background(), 1, 2, 3, state); err != nil {
		t.Fatalf("StoreDesiredAWDReconcileState() error = %v", err)
	}

	ttl := redisServer.TTL(rediskeys.DesiredAWDReconcileStateKey(1, 2, 3))
	if ttl <= 24*time.Hour {
		t.Fatalf("expected ttl to extend beyond retention floor, got %s", ttl)
	}
	if ttl > 24*time.Hour+10*time.Minute {
		t.Fatalf("expected ttl close to retry window plus retention, got %s", ttl)
	}
}
