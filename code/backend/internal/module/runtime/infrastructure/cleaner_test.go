package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/module/runtime/infrastructure/cachekeys"
)

type blockingCleanerService struct {
	started chan struct{}
	done    chan struct{}
}

func (s *blockingCleanerService) CleanExpiredInstances(ctx context.Context) error {
	close(s.started)
	<-ctx.Done()
	close(s.done)
	return ctx.Err()
}

func (s *blockingCleanerService) ReconcileLostActiveRuntimes(context.Context) error {
	return nil
}

func (s *blockingCleanerService) CleanupOrphans(context.Context) error {
	return nil
}

func TestCleanerStopCancelsRunningTask(t *testing.T) {
	t.Parallel()

	service := &blockingCleanerService{
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}
	cleaner := NewCleaner(service, nil, time.Minute, zap.NewNop())
	cleaner.baseCtx, cleaner.cancel = context.WithCancel(context.Background())

	cleaner.startRunOnce()

	select {
	case <-service.started:
	case <-time.After(time.Second):
		t.Fatal("cleaner task did not start")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cleaner.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	select {
	case <-service.done:
	case <-time.After(time.Second):
		t.Fatal("cleaner task did not stop after cancellation")
	}
}

func TestCleanerStopsRunningTaskWhenCleanupLockIsLost(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := &blockingCleanerService{
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}
	cleaner := NewCleaner(service, redisClient, 60*time.Millisecond, zap.NewNop())
	cleaner.baseCtx, cleaner.cancel = context.WithCancel(context.Background())
	t.Cleanup(cleaner.cancel)

	cleaner.startRunOnce()

	select {
	case <-service.started:
	case <-time.After(time.Second):
		t.Fatal("cleaner task did not start")
	}

	lockKey := cachekeys.ContainerCleanupLockKey()
	if !mini.Exists(lockKey) {
		t.Fatal("expected cleanup lock to exist after cleaner run starts")
	}
	if err := mini.Set(lockKey, "other-owner"); err != nil {
		t.Fatalf("replace cleanup lock token: %v", err)
	}
	mini.SetTTL(lockKey, time.Minute)

	select {
	case <-service.done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected cleaner task context to stop after cleanup lock is lost")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cleaner.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := &blockingCleanerService{
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}
	cleaner := NewCleaner(service, redisClient, time.Minute, zap.NewNop())
	cleaner.baseCtx, cleaner.cancel = context.WithCancel(context.Background())
	t.Cleanup(cleaner.cancel)

	cleaner.startRunOnce()

	select {
	case <-service.started:
	case <-time.After(time.Second):
		t.Fatal("cleaner task did not start")
	}

	lockKey := cachekeys.ContainerCleanupLockKey()
	if !mini.Exists(lockKey) {
		t.Fatal("expected cleanup lock to exist after cleaner run starts")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cleaner.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if mini.Exists(lockKey) {
		t.Fatal("expected cleanup lock to be released after cleaner shutdown")
	}
}

func TestStoppingCleanupLockStoreSkipsCallbackWhenLockHeld(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	lockKey := cachekeys.StoppingCleanupLockKey()
	if err := mini.Set(lockKey, "other-owner"); err != nil {
		t.Fatalf("seed stopping cleanup lock: %v", err)
	}
	mini.SetTTL(lockKey, time.Minute)

	store := NewStoppingCleanupLockStore(redisClient, time.Minute, zap.NewNop())
	called := false
	acquired, err := store.WithStoppingCleanupLock(context.Background(), func(context.Context) {
		called = true
	})
	if err != nil {
		t.Fatalf("WithStoppingCleanupLock() error = %v", err)
	}
	if acquired {
		t.Fatal("expected lock not to be acquired")
	}
	if called {
		t.Fatal("expected callback to be skipped when stopping cleanup lock is held")
	}
}

func TestStoppingCleanupLockStoreRunsCallbackAndReleasesLock(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	store := NewStoppingCleanupLockStore(redisClient, time.Minute, zap.NewNop())
	called := false
	acquired, err := store.WithStoppingCleanupLock(context.Background(), func(context.Context) {
		called = true
		if !mini.Exists(cachekeys.StoppingCleanupLockKey()) {
			t.Fatal("expected stopping cleanup lock to exist during callback")
		}
	})
	if err != nil {
		t.Fatalf("WithStoppingCleanupLock() error = %v", err)
	}
	if !acquired {
		t.Fatal("expected lock to be acquired")
	}
	if !called {
		t.Fatal("expected callback to run")
	}
	if mini.Exists(cachekeys.StoppingCleanupLockKey()) {
		t.Fatal("expected stopping cleanup lock to be released after callback")
	}
}

func TestCleanerStopRejectsNilContext(t *testing.T) {
	t.Parallel()

	cleaner := NewCleaner(&blockingCleanerService{}, nil, time.Minute, zap.NewNop())

	if err := cleaner.Stop(nil); err == nil {
		t.Fatal("expected Stop(nil) to reject missing context")
	}
}
