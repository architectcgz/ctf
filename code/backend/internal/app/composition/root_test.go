package composition

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
)

func TestRootProvidesEventsBus(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}
	if root == nil {
		t.Fatal("expected root")
	}
	if root.Events == nil {
		t.Fatal("expected events bus")
	}
}

func TestBackgroundJobStartDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	called := false
	job := NewBackgroundJob("ctx-check", func(ctx context.Context) error {
		called = true
		if ctx != nil {
			t.Fatalf("expected start ctx to stay nil, got %v", ctx)
		}
		return nil
	}, nil)

	if err := job.Start(nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !called {
		t.Fatal("expected start function to be called")
	}
}

func TestBackgroundJobStopDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	called := false
	job := NewBackgroundJob("ctx-check", nil, func(ctx context.Context) error {
		called = true
		if ctx != nil {
			t.Fatalf("expected stop ctx to stay nil, got %v", ctx)
		}
		return nil
	})

	if err := job.Stop(nil); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !called {
		t.Fatal("expected stop function to be called")
	}
}

func newRootTestDependencies(t *testing.T) (*config.Config, *gorm.DB, *redislib.Client) {
	t.Helper()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = cache.Close()
	})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "root.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	cfg := &config.Config{
		App: config.AppConfig{
			Name: "ctf-platform-test",
			Env:  "test",
		},
	}

	return cfg, db, cache
}
