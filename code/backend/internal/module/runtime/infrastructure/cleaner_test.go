package infrastructure

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
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

	go cleaner.runOnce()

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
