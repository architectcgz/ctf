package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app"
	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/postgres"
	infraredis "ctf-platform/internal/infrastructure/redis"
	"ctf-platform/pkg/logger"
)

func Run() {
	rootCtx := context.Background()
	env := os.Getenv("APP_ENV")
	cfg, err := config.Load(env)
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	db := mustOpenPostgres(rootCtx, cfg, log)
	cache := mustOpenRedis(rootCtx, cfg, log)

	server, err := app.NewHTTPServer(cfg, log, db, cache)
	if err != nil {
		closeResources(log, db, cache)
		log.Fatal("http_server_init_failed", zap.Error(err))
	}

	go func() {
		log.Info("http_server_starting",
			zap.String("env", cfg.App.Env),
			zap.String("addr", cfg.HTTP.Host),
			zap.Int("port", cfg.HTTP.Port),
		)

		if serveErr := server.Start(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Fatal("http_server_failed", zap.Error(serveErr))
		}
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := shutdownGracefully(signalCtx, stop, server, 10*time.Second); err != nil {
		log.Error("http_server_shutdown_failed", zap.Error(err))
	} else {
		log.Info("http_server_stopped")
	}
	closeResources(log, db, cache)
}

func mustOpenPostgres(ctx context.Context, cfg *config.Config, log *zap.Logger) *gorm.DB {
	db, err := postgres.Open(ctx, cfg.Postgres)
	if err != nil {
		log.Fatal("postgres_init_failed", zap.Error(err))
	}
	return db
}

func mustOpenRedis(ctx context.Context, cfg *config.Config, log *zap.Logger) *redislib.Client {
	client, err := infraredis.NewClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatal("redis_init_failed", zap.Error(err))
	}
	return client
}

type shutdownServer interface {
	Shutdown(ctx context.Context) error
}

func shutdownGracefully(waitCtx context.Context, stop func(), server shutdownServer, timeout time.Duration) error {
	if waitCtx == nil {
		return errors.New("shutdown wait context is required")
	}
	if server == nil {
		return errors.New("shutdown server is required")
	}

	<-waitCtx.Done()
	if stop != nil {
		stop()
	}

	shutdownCtx := context.Background()
	cancel := func() {}
	if timeout > 0 {
		shutdownCtx, cancel = context.WithTimeout(context.Background(), timeout)
	}
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

func closeResources(log *zap.Logger, db *gorm.DB, cache *redislib.Client) {
	closePostgres(log, db)
	closeRedis(log, cache)
}

func closePostgres(log *zap.Logger, db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Warn("postgres_sql_db_unavailable_for_close", zap.Error(err))
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Warn("postgres_close_failed", zap.Error(err))
		return
	}
	log.Info("postgres_closed")
}

func closeRedis(log *zap.Logger, cache *redislib.Client) {
	if cache == nil {
		return
	}
	if err := cache.Close(); err != nil {
		log.Warn("redis_close_failed", zap.Error(err))
		return
	}
	log.Info("redis_closed")
}
