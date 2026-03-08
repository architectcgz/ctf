package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewRouterRegistersStudentChallengeRoutes(t *testing.T) {
	t.Setenv("CTF_FLAG_SECRET", "12345678901234567890123456789012")

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

	dbPath := filepath.Join(t.TempDir(), "router.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	router, err := NewRouter(newPracticeFlowTestConfig(t), zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	assertHasRoute(t, router, "GET", "/api/v1/challenges")
	assertHasRoute(t, router, "GET", "/api/v1/challenges/:id")
}

func assertHasRoute(t *testing.T, router *gin.Engine, method, path string) {
	t.Helper()

	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			return
		}
	}

	t.Fatalf("route not found: %s %s", method, path)
}
