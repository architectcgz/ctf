package app

import (
	"context"
	"fmt"
	"net/http"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/container"
	"ctf-platform/internal/module/contest"
)

type HTTPServer struct {
	server        *http.Server
	cleaner       *container.Cleaner
	statusUpdater *contest.StatusUpdater
	updaterCtx    context.Context
	updaterCancel context.CancelFunc
	logger        *zap.Logger
}

func NewHTTPServer(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*HTTPServer, error) {
	engine, err := NewRouter(cfg, log, db, cache)
	if err != nil {
		return nil, err
	}

	containerRepo := container.NewRepository(db)
	containerService := container.NewService(containerRepo, &cfg.Container, log.Named("container_service"))
	cleaner := container.NewCleaner(containerService, log.Named("container_cleaner"))

	if err := cleaner.Start(cfg.Container.CleanupInterval); err != nil {
		return nil, fmt.Errorf("启动清理任务失败: %w", err)
	}

	contestRepo := contest.NewRepository(db)
	statusUpdater := contest.NewStatusUpdater(
		contestRepo,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		log.Named("contest_status_updater"),
	)
	updaterCtx, updaterCancel := context.WithCancel(context.Background())
	go statusUpdater.Start(updaterCtx)

	return &HTTPServer{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
			Handler:      engine,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
		cleaner:       cleaner,
		statusUpdater: statusUpdater,
		updaterCtx:    updaterCtx,
		updaterCancel: updaterCancel,
		logger:        log,
	}, nil
}

func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("停止清理任务")
	s.cleaner.Stop()
	s.logger.Info("停止竞赛状态更新任务")
	s.updaterCancel()
	return s.server.Shutdown(ctx)
}
