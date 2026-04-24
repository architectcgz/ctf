package queries

import (
	"context"
	"testing"
)

type runtimeCountRunningContextKey string

type stubCountRunningRepository struct {
	countRunningWithContextFn func(ctx context.Context) (int64, error)
}

func (s *stubCountRunningRepository) CountRunning(ctx context.Context) (int64, error) {
	if s.countRunningWithContextFn != nil {
		return s.countRunningWithContextFn(ctx)
	}
	return 0, nil
}

func TestCountRunningServiceCountRunningPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := runtimeCountRunningContextKey("count-running")
	expectedCtxValue := "ctx-count-running"
	called := false
	service := NewCountRunningService(&stubCountRunningRepository{
		countRunningWithContextFn: func(ctx context.Context) (int64, error) {
			called = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected count-running ctx value %v, got %v", expectedCtxValue, got)
			}
			return 7, nil
		},
	})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	count, err := service.CountRunning(ctx)
	if err != nil {
		t.Fatalf("CountRunning() error = %v", err)
	}
	if !called {
		t.Fatal("expected context-aware count running repository to be called")
	}
	if count != 7 {
		t.Fatalf("CountRunning() count = %d, want 7", count)
	}
}
