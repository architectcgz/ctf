package lockkeepalive

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeLease struct {
	mu        sync.Mutex
	refreshFn func(context.Context, time.Duration) (bool, error)
}

func (l *fakeLease) Refresh(ctx context.Context, ttl time.Duration) (bool, error) {
	l.mu.Lock()
	fn := l.refreshFn
	l.mu.Unlock()
	if fn == nil {
		return true, nil
	}
	return fn(ctx, ttl)
}

func (l *fakeLease) Key(context.Context) string { return "fake-lock" }

func TestStartCancelsRunContextWhenLeaseRefreshReturnsFalse(t *testing.T) {
	t.Parallel()

	lease := &fakeLease{
		refreshFn: func(context.Context, time.Duration) (bool, error) {
			return false, nil
		},
	}

	runCtx, stop := Start(context.Background(), nil, lease, "test-lock", 30*time.Millisecond)
	defer stop()

	select {
	case <-runCtx.Done():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected run context to be canceled after lease refresh returned false")
	}
}

func TestRefreshIntervalMatchesKeepalivePolicy(t *testing.T) {
	t.Parallel()

	if got := RefreshInterval(30 * time.Second); got != 10*time.Second {
		t.Fatalf("expected 30s ttl refresh interval to be 10s, got %s", got)
	}
	if got := RefreshInterval(3 * time.Second); got != 1500*time.Millisecond {
		t.Fatalf("expected 3s ttl refresh interval to be 1.5s, got %s", got)
	}
}

func TestFailoverWindowAccountsForTTLRefreshAndRetry(t *testing.T) {
	t.Parallel()

	got := FailoverWindow(40*time.Second, time.Second)
	want := 40*time.Second + (40 * time.Second / 3) + time.Second
	if got != want {
		t.Fatalf("expected failover window %s, got %s", want, got)
	}
}

func TestStartCancelsRunContextAfterRefreshErrorsExceedTTL(t *testing.T) {
	t.Parallel()

	lease := &fakeLease{
		refreshFn: func(context.Context, time.Duration) (bool, error) {
			return false, errors.New("redis unavailable")
		},
	}

	runCtx, stop := Start(context.Background(), nil, lease, "test-lock", 40*time.Millisecond)
	defer stop()

	select {
	case <-runCtx.Done():
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected run context to be canceled after refresh errors exceeded ttl")
	}
}

func TestStartCancelsRunContextWhenRefreshBlocksPastTTL(t *testing.T) {
	t.Parallel()

	lease := &fakeLease{
		refreshFn: func(ctx context.Context, _ time.Duration) (bool, error) {
			<-ctx.Done()
			return false, ctx.Err()
		},
	}

	runCtx, stop := Start(context.Background(), nil, lease, "test-lock", 60*time.Millisecond)
	defer stop()

	select {
	case <-runCtx.Done():
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected run context to be canceled after refresh blocked past ttl")
	}
}
