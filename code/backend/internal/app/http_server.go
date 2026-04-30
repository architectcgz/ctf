package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
)

type lifecycleCloser interface {
	Close(ctx context.Context) error
}

type lifecycleComponent struct {
	name   string
	closer lifecycleCloser
}

type HTTPServer struct {
	server         *http.Server
	backgroundJobs []composition.BackgroundJob
	closers        []lifecycleComponent
	appCtx         context.Context
	cancelApp      context.CancelFunc
	logger         *zap.Logger
}

func NewHTTPServer(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*HTTPServer, error) {
	root, err := composition.BuildRoot(cfg, log, db, cache)
	if err != nil {
		return nil, err
	}

	routerRuntime, err := buildRouterRuntime(root)
	if err != nil {
		return nil, err
	}
	server := &HTTPServer{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
			Handler:      routerRuntime.engine,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
		backgroundJobs: root.BackgroundJobs(),
		closers:        routerRuntime.closers,
		appCtx:         root.Context(),
		cancelApp:      root.Cancel,
		logger:         log,
	}
	if err := server.startBackgroundJobs(); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	var shutdownErrs []error

	if s.cancelApp != nil {
		s.cancelApp()
	}

	if err := s.stopBackgroundJobs(ctx); err != nil {
		shutdownErrs = append(shutdownErrs, err)
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

func (s *HTTPServer) startBackgroundJobs() error {
	ctx := s.appCtx
	if ctx == nil {
		return errors.New("http server background jobs require application context")
	}
	for _, job := range s.backgroundJobs {
		s.logger.Info("启动后台任务", zap.String("job", job.Name()))
		if err := job.Start(ctx); err != nil {
			return fmt.Errorf("%s: %w", job.Name(), err)
		}
	}
	return nil
}

func (s *HTTPServer) stopBackgroundJobs(ctx context.Context) error {
	var errs []error
	for _, job := range s.backgroundJobs {
		s.logger.Info("停止后台任务", zap.String("job", job.Name()))
		if err := job.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", job.Name(), err))
		}
	}
	return errors.Join(errs...)
}
