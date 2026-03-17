package bootstrap

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
