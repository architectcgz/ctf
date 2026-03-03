package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/app"
	"ctf-platform/internal/config"
	"ctf-platform/pkg/logger"
)

func main() {
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

	server := app.NewHTTPServer(cfg, log)

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
