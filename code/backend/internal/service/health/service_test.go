package health

import (
	"context"
	"errors"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
)

func TestCheckLiveDoesNotRequireDependencies(t *testing.T) {
	t.Parallel()

	svc := NewService(testHealthConfig(), nil, nil, NewReadinessState())

	status := svc.CheckLive(context.Background())
	if status.HTTPStatus() != 200 {
		t.Fatalf("CheckLive() HTTPStatus = %d, want 200", status.HTTPStatus())
	}
	if status.HealthStatus.Status != "ok" {
		t.Fatalf("CheckLive() status = %q, want ok", status.HealthStatus.Status)
	}
}

func TestCheckReadyFailsWhenProcessIsDraining(t *testing.T) {
	t.Parallel()

	db := newHealthTestDB(t)
	redis := newHealthTestRedis(t)
	readiness := NewReadinessState()
	svc := NewService(testHealthConfig(), db, redis, readiness)

	ready := svc.CheckReady(context.Background())
	if ready.HTTPStatus() != 200 {
		t.Fatalf("CheckReady() HTTPStatus = %d, want 200", ready.HTTPStatus())
	}

	readiness.MarkDraining()
	draining := svc.CheckReady(context.Background())
	if draining.HTTPStatus() != 503 {
		t.Fatalf("draining CheckReady() HTTPStatus = %d, want 503", draining.HTTPStatus())
	}
	if draining.HealthStatus.Status != "not_ready" {
		t.Fatalf("draining CheckReady() status = %q, want not_ready", draining.HealthStatus.Status)
	}
	if draining.HealthStatus.Dependencies["process"] != "draining" {
		t.Fatalf("process dependency = %q, want draining", draining.HealthStatus.Dependencies["process"])
	}
}

func TestCheckReadyFailsWhenDependencyIsDown(t *testing.T) {
	t.Parallel()

	db := newHealthTestDB(t)
	redis := newHealthTestRedis(t)
	_ = redis.Close()
	svc := NewService(testHealthConfig(), db, redis, NewReadinessState())

	status := svc.CheckReady(context.Background())
	if status.HTTPStatus() != 503 {
		t.Fatalf("CheckReady() HTTPStatus = %d, want 503", status.HTTPStatus())
	}
	if status.HealthStatus.Dependencies["redis"] != "down" {
		t.Fatalf("redis dependency = %q, want down", status.HealthStatus.Dependencies["redis"])
	}
}

func TestCheckReadyFailsWhenContainerFlagSecretCheckFails(t *testing.T) {
	t.Parallel()

	db := newHealthTestDB(t)
	redis := newHealthTestRedis(t)
	svc := NewService(testHealthConfig(), db, redis, NewReadinessState(), DependencyCheck{
		Name: "container_flag_secret",
		Check: func(context.Context) error {
			return errors.New("fingerprint mismatch")
		},
	})

	status := svc.CheckReady(context.Background())
	if status.HTTPStatus() != 503 {
		t.Fatalf("CheckReady() HTTPStatus = %d, want 503", status.HTTPStatus())
	}
	if status.HealthStatus.Dependencies["container_flag_secret"] != "down" {
		t.Fatalf("container flag secret dependency = %q, want down", status.HealthStatus.Dependencies["container_flag_secret"])
	}
}

func testHealthConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:    "ctf-platform",
			Env:     "test",
			Version: "test",
		},
	}
}

func newHealthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func newHealthTestRedis(t *testing.T) *redislib.Client {
	t.Helper()

	mini := miniredis.RunT(t)
	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	return client
}
