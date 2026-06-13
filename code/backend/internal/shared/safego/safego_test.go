package safego_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"ctf-platform/internal/shared/safego"
)

func TestGoRecoversPanicAndLogsStack(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core).With(zap.String("request_id", "req-safego-1"))
	ctx := context.Background()

	var wg sync.WaitGroup
	started := make(chan struct{})

	safego.Go(&wg, ctx, logger, "practice_async_task", func(context.Context) {
		close(started)
		panic("boom")
	})

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected safego task to start")
	}

	wg.Wait()

	if observed.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", observed.Len())
	}
	entry := observed.All()[0]
	if entry.Message != "panic_recovered" {
		t.Fatalf("message = %q, want panic_recovered", entry.Message)
	}
	fields := entry.ContextMap()
	if got := fields["task_name"]; got != "practice_async_task" {
		t.Fatalf("task_name = %v, want practice_async_task", got)
	}
	if got := fields["request_id"]; got != "req-safego-1" {
		t.Fatalf("request_id = %v, want req-safego-1", got)
	}
	if got := fields["panic"]; got != "boom" {
		t.Fatalf("panic = %v, want boom", got)
	}
	if _, ok := fields["stack"]; !ok {
		t.Fatal("expected stack field in panic log")
	}
}

func TestGoUsesBackgroundContextWhenContextNil(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	received := make(chan context.Context, 1)

	safego.Go(&wg, nil, nil, "nil_context_task", func(ctx context.Context) {
		received <- ctx
	})

	wg.Wait()

	select {
	case got := <-received:
		if got == nil {
			t.Fatal("expected safego to provide a non-nil background context")
		}
		if err := got.Err(); err != nil {
			t.Fatalf("expected background context to stay active, got %v", err)
		}
	default:
		t.Fatal("expected task to receive a context")
	}
}

func TestGoRecoversPanicWithNilLogger(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	started := make(chan struct{})

	safego.Go(&wg, nil, nil, "nil_logger_task", func(context.Context) {
		close(started)
		panic("boom")
	})

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected safego task to start")
	}

	wg.Wait()
}
