package composition

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLoopBackgroundJobLogsPanicThroughSafeGo(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	job := NewLoopBackgroundJobWithLogger("root_async_job", logger, func(context.Context) {
		panic("boom")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := job.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	deadline := time.After(time.Second)
	for observed.Len() == 0 {
		select {
		case <-deadline:
			t.Fatal("expected panic recovery log entry")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := job.Stop(stopCtx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	entry := observed.All()[0]
	if entry.Message != "panic_recovered" {
		t.Fatalf("message = %q, want panic_recovered", entry.Message)
	}
	fields := entry.ContextMap()
	if got := fields["task_name"]; got != "root_async_job" {
		t.Fatalf("task_name = %v, want root_async_job", got)
	}
	if _, ok := fields["stack"]; !ok {
		t.Fatal("expected stack field in panic log")
	}
}

func TestBackgroundJobStartUsesProvidedContext(t *testing.T) {
	t.Parallel()

	var got context.Context
	var mu sync.Mutex
	job := NewBackgroundJob("context_job", func(ctx context.Context) error {
		mu.Lock()
		got = ctx
		mu.Unlock()
		return nil
	}, nil)

	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	if err := job.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if got != ctx {
		t.Fatalf("start ctx = %v, want original context", got)
	}
}

func TestNewLoopBackgroundJobRejectsNilContext(t *testing.T) {
	t.Parallel()

	job := NewLoopBackgroundJob("nil_ctx_job", func(context.Context) {})
	if err := job.Start(nil); err == nil {
		t.Fatal("expected nil context to be rejected")
	}
}
