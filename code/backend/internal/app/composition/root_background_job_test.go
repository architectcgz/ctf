package composition

import (
	"context"
	"sync"
	"testing"
)

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
