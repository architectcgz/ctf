package app

import (
	"context"
	"fmt"
	"net/http"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/assessment"
	"ctf-platform/internal/module/container"
	"ctf-platform/internal/module/contest"
)

type HTTPServer struct {
	server        *http.Server
	cleaner       *container.Cleaner
	assessment    *assessment.Cleaner
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
	var containerEngine *container.Engine
	if engine, err := container.NewEngine(&cfg.Container); err != nil {
		log.Warn("container_engine_init_failed_for_cleaner", zap.Error(err))
	} else {
		containerEngine = engine
	}
	containerService := container.NewService(containerRepo, containerEngine, &cfg.Container, log.Named("container_service"))
	cleaner := container.NewCleaner(containerService, log.Named("container_cleaner"))

	if err := cleaner.Start(cfg.Container.CleanupInterval); err != nil {
		return nil, fmt.Errorf("启动清理任务失败: %w", err)
	}

	assessmentRepo := assessment.NewRepository(db)
	assessmentService := assessment.NewService(assessmentRepo, cache, cfg.Assessment, log.Named("assessment_service"))
	assessmentCleaner := assessment.NewCleaner(assessmentService, log.Named("assessment_cleaner"))
	if err := assessmentCleaner.Start(cfg.Assessment.FullRebuildCron, cfg.Assessment.FullRebuildTimeout); err != nil {
		return nil, fmt.Errorf("启动能力画像任务失败: %w", err)
	}

	contestRepo := contest.NewRepository(db)
	statusUpdater := contest.NewStatusUpdater(
		contestRepo,
		cache,
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
		assessment:    assessmentCleaner,
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
	s.logger.Info("停止能力画像任务")
	s.assessment.Stop()
	s.logger.Info("停止竞赛状态更新任务")
	s.updaterCancel()
	return s.server.Shutdown(ctx)
}
