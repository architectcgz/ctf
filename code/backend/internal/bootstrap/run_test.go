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

type stubShutdownServer struct {
	shutdown func(context.Context) error
}

func (s stubShutdownServer) Shutdown(ctx context.Context) error {
	return s.shutdown(ctx)
}

func TestClosePostgresClosesSQLDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	closePostgres(zap.NewNop(), db)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	if err := sqlDB.PingContext(context.Background()); err == nil {
		t.Fatal("expected closed sql db ping to fail")
	}
}

func TestCloseRedisClosesClient(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})

	closeRedis(zap.NewNop(), client)

	if err := client.Ping(context.Background()).Err(); err == nil {
		t.Fatal("expected closed redis client ping to fail")
	}
}

func TestShutdownGracefullyUsesTimeoutContextAndReleasesSignalHandler(t *testing.T) {
	t.Parallel()

	waitCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopCalled := make(chan struct{}, 1)
	server := stubShutdownServer{
		shutdown: func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("expected shutdown context deadline")
			}
			if time.Until(deadline) <= 0 {
				t.Fatal("expected shutdown context deadline to be in the future")
			}
			return nil
		},
	}

	done := make(chan error, 1)
	go func() {
		done <- shutdownGracefully(waitCtx, func() {
			stopCalled <- struct{}{}
		}, server, 200*time.Millisecond)
	}()

	cancel()

	select {
	case <-stopCalled:
	case <-time.After(time.Second):
		t.Fatal("expected signal stop func to be called")
	}

	if err := <-done; err != nil {
		t.Fatalf("shutdownGracefully() error = %v", err)
	}
}

func TestShutdownGracefullyReturnsServerShutdownError(t *testing.T) {
	t.Parallel()

	waitCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedErr := errors.New("shutdown failed")
	server := stubShutdownServer{
		shutdown: func(context.Context) error {
			return expectedErr
		},
	}

	done := make(chan error, 1)
	go func() {
		done <- shutdownGracefully(waitCtx, func() {}, server, time.Second)
	}()

	cancel()

	if err := <-done; !errors.Is(err, expectedErr) {
		t.Fatalf("shutdownGracefully() error = %v, want %v", err, expectedErr)
	}
}
