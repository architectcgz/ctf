package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func TestAWDRoundUpdaterRefreshesSchedulerLockWhileRunning(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	updater := &AWDRoundUpdater{
		redis: redisClient,
		cfg: config.ContestAWDConfig{
			SchedulerLockTTL: 60 * time.Millisecond,
		},
		log: zap.NewNop(),
	}

	lockKey := rediskeys.AWDSchedulerLockKey()
	updater.withSchedulerLock(context.Background(), func(ctx context.Context) {
		time.Sleep(140 * time.Millisecond)
		if ctx.Err() != nil {
			t.Fatalf("expected scheduler lock context to remain active, got %v", ctx.Err())
		}
		if !mini.Exists(lockKey) {
			t.Fatalf("expected scheduler lock %q to be refreshed during run", lockKey)
		}
		if ttl := mini.TTL(lockKey); ttl <= 0 {
			t.Fatalf("expected positive scheduler lock ttl during run, got %s", ttl)
		}
	})

	if mini.Exists(lockKey) {
		t.Fatalf("expected scheduler lock %q to be released after run", lockKey)
	}
}
