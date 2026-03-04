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

	db := mustOpenPostgres(cfg, log)
	cache := mustOpenRedis(cfg, log)

	server, err := app.NewHTTPServer(cfg, log, db, cache)
	if err != nil {
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

	waitForShutdown(log, server)
}

func mustOpenPostgres(cfg *config.Config, log *zap.Logger) *gorm.DB {
	db, err := postgres.Open(cfg.Postgres)
	if err != nil {
		log.Fatal("postgres_init_failed", zap.Error(err))
	}
	return db
}

func mustOpenRedis(cfg *config.Config, log *zap.Logger) *redislib.Client {
	client, err := infraredis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal("redis_init_failed", zap.Error(err))
	}
	return client
}

func waitForShutdown(log *zap.Logger, server *app.HTTPServer) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("http_server_shutdown_failed", zap.Error(err))
		return
	}

	log.Info("http_server_stopped")
}
