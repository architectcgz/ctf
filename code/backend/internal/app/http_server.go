package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/assessment"
	"ctf-platform/internal/module/container"
	"ctf-platform/internal/module/contest"
)

type reportServiceCloser interface {
	Close(ctx context.Context) error
}

type lifecycleComponent struct {
	name   string
	closer reportServiceCloser
}

type HTTPServer struct {
	server        *http.Server
	cleaner       *container.Cleaner
	assessment    *assessment.Cleaner
	closers       []lifecycleComponent
	statusUpdater *contest.StatusUpdater
	awdUpdater    *contest.AWDRoundUpdater
	updaterCancel context.CancelFunc
	updaterWG     *sync.WaitGroup
	logger        *zap.Logger
}

func NewHTTPServer(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*HTTPServer, error) {
	routerRuntime, err := buildRouterRuntime(cfg, log, db, cache)
	if err != nil {
		return nil, err
	}
	if routerRuntime.containerService == nil {
		return nil, fmt.Errorf("container service not initialized")
	}
	if routerRuntime.assessmentService == nil {
		return nil, fmt.Errorf("assessment service not initialized")
	}

	cleaner := container.NewCleaner(routerRuntime.containerService, cache, cfg.Container.CleanupLockTTL, log.Named("container_cleaner"))

	if err := cleaner.Start(cfg.Container.CleanupInterval); err != nil {
		return nil, fmt.Errorf("启动清理任务失败: %w", err)
	}

	assessmentCleaner := assessment.NewCleaner(routerRuntime.assessmentService, log.Named("assessment_cleaner"))
	if err := assessmentCleaner.Start(cfg.Assessment.FullRebuildCron, cfg.Assessment.FullRebuildTimeout); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cleaner.Stop(cleanupCtx)
		return nil, fmt.Errorf("启动能力画像任务失败: %w", err)
	}

	contestRepo := contest.NewRepository(db)
	statusUpdater := contest.NewStatusUpdater(
		contestRepo,
		cache,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		cfg.Contest.StatusUpdateLockTTL,
		log.Named("contest_status_updater"),
	)
	awdUpdater := contest.NewAWDRoundUpdater(
		db,
		cache,
		cfg.Contest.AWD,
		cfg.Container.FlagGlobalSecret,
		contest.NewDockerAWDFlagInjector(db, routerRuntime.containerService, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
	)
	updaterCtx, updaterCancel := context.WithCancel(context.Background())
	updaterWG := &sync.WaitGroup{}
	updaterWG.Add(2)
	go func() {
		defer updaterWG.Done()
		statusUpdater.Start(updaterCtx)
	}()
	go func() {
		defer updaterWG.Done()
		awdUpdater.Start(updaterCtx)
	}()

	return &HTTPServer{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
			Handler:      routerRuntime.engine,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
		cleaner:       cleaner,
		assessment:    assessmentCleaner,
		closers:       routerRuntime.closers,
		statusUpdater: statusUpdater,
		awdUpdater:    awdUpdater,
		updaterCancel: updaterCancel,
		updaterWG:     updaterWG,
		logger:        log,
	}, nil
}

func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	var shutdownErrs []error

	s.logger.Info("停止竞赛状态更新任务")
	s.logger.Info("停止 AWD 轮次推进任务")
	if s.updaterCancel != nil {
		s.updaterCancel()
	}
	if err := s.waitForUpdaters(ctx); err != nil {
		shutdownErrs = append(shutdownErrs, err)
	}

	s.logger.Info("停止清理任务")
	if s.cleaner != nil {
		if err := s.cleaner.Stop(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, err)
		}
	}

	s.logger.Info("停止能力画像任务")
	if s.assessment != nil {
		if err := s.assessment.Stop(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, err)
		}
	}

	for _, component := range s.closers {
		if component.closer == nil {
			continue
		}
		s.logger.Info("停止应用异步任务", zap.String("component", component.name))
		if err := component.closer.Close(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, fmt.Errorf("%s: %w", component.name, err))
		}
	}

	if err := s.server.Shutdown(ctx); err != nil {
		shutdownErrs = append(shutdownErrs, err)
	}
	return errors.Join(shutdownErrs...)
}

func (s *HTTPServer) waitForUpdaters(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		if s.updaterWG != nil {
			s.updaterWG.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
