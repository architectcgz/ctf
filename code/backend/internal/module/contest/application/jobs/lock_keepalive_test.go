package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/pkg/redislock"
)

func TestRedisLockKeepaliveRefreshesOwnedLock(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	lock, acquired, err := redislock.Acquire(context.Background(), client, "contest:status:keepalive", 60*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if !acquired {
		t.Fatal("expected lock acquisition")
	}

	runCtx, stop := startRedisLockKeepalive(context.Background(), zap.NewNop(), lock, redisLockKeepaliveConfig{
		Name: "contest_status_updater",
		TTL:  60 * time.Millisecond,
	})
	defer stop()

	time.Sleep(140 * time.Millisecond)
	if runCtx.Err() != nil {
		t.Fatalf("expected keepalive context to remain active, got %v", runCtx.Err())
	}
	if !mini.Exists("contest:status:keepalive") {
		t.Fatal("expected lock key to exist while keepalive is running")
	}
	if ttl := mini.TTL("contest:status:keepalive"); ttl <= 0 {
		t.Fatalf("expected positive ttl during keepalive, got %s", ttl)
	}
}

func TestRedisLockKeepaliveCancelsWhenTokenChanges(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	lock, acquired, err := redislock.Acquire(context.Background(), client, "contest:status:lost", 60*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if !acquired {
		t.Fatal("expected lock acquisition")
	}

	runCtx, stop := startRedisLockKeepalive(context.Background(), zap.NewNop(), lock, redisLockKeepaliveConfig{
		Name: "contest_status_updater",
		TTL:  60 * time.Millisecond,
	})
	defer stop()

	if err := mini.Set("contest:status:lost", "other-token"); err != nil {
		t.Fatalf("replace token: %v", err)
	}
	mini.SetTTL("contest:status:lost", time.Minute)

	select {
	case <-runCtx.Done():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected run context to cancel after token loss")
	}
}
