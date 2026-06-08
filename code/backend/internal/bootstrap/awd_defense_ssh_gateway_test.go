package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubAWDDefenseSSHGatewayRunner struct {
	startCalls int
	stopCalls  int
	startErr   error
	stopErr    error
	onStop     func()
}

func (s *stubAWDDefenseSSHGatewayRunner) Start(context.Context) error {
	s.startCalls++
	return s.startErr
}

func (s *stubAWDDefenseSSHGatewayRunner) Stop(context.Context) error {
	s.stopCalls++
	if s.onStop != nil {
		s.onStop()
	}
	return s.stopErr
}

type stubAWDDefenseSSHGatewayRuntimeCloser struct {
	closeCalls int
	closeErr   error
}

func (s *stubAWDDefenseSSHGatewayRuntimeCloser) Close(context.Context) error {
	s.closeCalls++
	return s.closeErr
}

func TestRunAWDDefenseSSHGatewayProcessStartPropagatesRunnerError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("start failed")
	process := &awdDefenseSSHGatewayProcess{
		gateway: &stubAWDDefenseSSHGatewayRunner{startErr: expectedErr},
	}

	if err := process.Start(context.Background()); !errors.Is(err, expectedErr) {
		t.Fatalf("Start() error = %v, want %v", err, expectedErr)
	}
}

func TestRunAWDDefenseSSHGatewayProcessShutdownStopsGatewayAndClosesResources(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	mini := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	runner := &stubAWDDefenseSSHGatewayRunner{}
	closer := &stubAWDDefenseSSHGatewayRuntimeCloser{}
	cancelled := false
	process := &awdDefenseSSHGatewayProcess{
		cancel: func() {
			cancelled = true
		},
		gateway:        runner,
		runtimeCloser:  closer,
		log:            zap.NewNop(),
		db:             db,
		cache:          cache,
		shutdownTimout: 250 * time.Millisecond,
	}

	if err := process.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if !cancelled {
		t.Fatal("expected shutdown to cancel root context")
	}
	if runner.stopCalls != 1 {
		t.Fatalf("expected gateway stop to be called once, got %d", runner.stopCalls)
	}
	if closer.closeCalls != 1 {
		t.Fatalf("expected runtime closer to be called once, got %d", closer.closeCalls)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	if err := sqlDB.PingContext(context.Background()); err == nil {
		t.Fatal("expected closed sql db ping to fail")
	}
	if err := cache.Ping(context.Background()).Err(); err == nil {
		t.Fatal("expected closed redis client ping to fail")
	}
}

func TestRunAWDDefenseSSHGatewayProcessShutdownCancelsBeforeStoppingGateway(t *testing.T) {
	t.Parallel()

	cancelled := false
	runner := &stubAWDDefenseSSHGatewayRunner{
		onStop: func() {
			if !cancelled {
				t.Fatal("expected root context to be cancelled before gateway Stop()")
			}
		},
	}
	process := &awdDefenseSSHGatewayProcess{
		cancel:  func() { cancelled = true },
		gateway: runner,
		log:     zap.NewNop(),
	}

	if err := process.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestRunAWDDefenseSSHGatewayProcessShutdownSkipsDownstreamClosersWhenGatewayStopFails(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	mini := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	runner := &stubAWDDefenseSSHGatewayRunner{stopErr: context.DeadlineExceeded}
	closer := &stubAWDDefenseSSHGatewayRuntimeCloser{}
	process := &awdDefenseSSHGatewayProcess{
		cancel:         func() {},
		gateway:        runner,
		runtimeCloser:  closer,
		log:            zap.NewNop(),
		db:             db,
		cache:          cache,
		shutdownTimout: 250 * time.Millisecond,
	}

	err = process.Shutdown(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown() error = %v, want %v", err, context.DeadlineExceeded)
	}
	if closer.closeCalls != 0 {
		t.Fatalf("expected runtime closer to be skipped after gateway stop failure, got %d calls", closer.closeCalls)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	if err := sqlDB.PingContext(context.Background()); err != nil {
		t.Fatalf("expected sql db to remain open after gateway stop failure, got %v", err)
	}
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("expected redis client to remain open after gateway stop failure, got %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		t.Fatalf("sqlDB.Close() error = %v", err)
	}
	if err := cache.Close(); err != nil {
		t.Fatalf("cache.Close() error = %v", err)
	}
}
