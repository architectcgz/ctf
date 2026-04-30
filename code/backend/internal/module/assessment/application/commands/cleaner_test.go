package commands

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

type blockingRebuildService struct {
	started chan struct{}
	done    chan struct{}
}

func (s *blockingRebuildService) RebuildAllSkillProfiles(ctx context.Context) error {
	close(s.started)
	<-ctx.Done()
	close(s.done)
	return ctx.Err()
}

func TestCleanerStopCancelsRunningRebuild(t *testing.T) {
	t.Parallel()

	service := &blockingRebuildService{
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}
	cleaner := NewCleaner(service, zap.NewNop())
	cleaner.baseCtx, cleaner.cancel = context.WithCancel(context.Background())

	go cleaner.runOnce(time.Minute)

	select {
	case <-service.started:
	case <-time.After(time.Second):
		t.Fatal("rebuild task did not start")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cleaner.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	select {
	case <-service.done:
	case <-time.After(time.Second):
		t.Fatal("rebuild task did not stop after cancellation")
	}
}

func TestCleanerStopRejectsNilContext(t *testing.T) {
	t.Parallel()

	cleaner := NewCleaner(&blockingRebuildService{
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}, zap.NewNop())

	if err := cleaner.Stop(nil); err == nil {
		t.Fatal("expected Stop(nil) to reject missing context")
	}
}
