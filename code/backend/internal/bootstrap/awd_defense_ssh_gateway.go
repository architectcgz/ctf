package bootstrap

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/logger"
)

type awdDefenseSSHGatewayRunner interface {
	Start(ctx context.Context) error
	Drain(ctx context.Context) error
	Stop(ctx context.Context) error
}

type awdDefenseSSHGatewayRuntimeCloser interface {
	Close(ctx context.Context) error
}

type awdDefenseSSHGatewayProcess struct {
	cancel         func()
	gateway        awdDefenseSSHGatewayRunner
	runtimeCloser  awdDefenseSSHGatewayRuntimeCloser
	log            *zap.Logger
	db             *gorm.DB
	cache          *redislib.Client
	shutdownTimout time.Duration
}

func RunAWDDefenseSSHGateway() {
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

	process, err := buildAWDDefenseSSHGatewayProcess(rootCtx, cfg, log)
	if err != nil {
		log.Fatal("awd_defense_ssh_gateway_init_failed", zap.Error(err))
	}
	if err := process.Start(rootCtx); err != nil {
		_ = process.Shutdown(rootCtx)
		log.Fatal("awd_defense_ssh_gateway_start_failed", zap.Error(err))
	}

	signalCtx, stop := signal.NotifyContext(rootCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := shutdownGracefully(signalCtx, stop, process, process.shutdownTimeout()); err != nil {
		log.Error("awd_defense_ssh_gateway_shutdown_failed", zap.Error(err))
	} else {
		log.Info("awd_defense_ssh_gateway_stopped")
	}
}

func buildAWDDefenseSSHGatewayProcess(ctx context.Context, cfg *config.Config, log *zap.Logger) (*awdDefenseSSHGatewayProcess, error) {
	if cfg == nil {
		return nil, errors.New("awd defense ssh gateway config is required")
	}
	if log == nil {
		log = zap.NewNop()
	}

	db := mustOpenPostgres(ctx, cfg, log)
	cache := mustOpenRedis(ctx, cfg, log)

	root, err := composition.BuildRoot(cfg, log, db, cache)
	if err != nil {
		closeResources(log, db, cache)
		return nil, err
	}
	runtime, err := composition.BuildContainerRuntimeModule(root)
	if err != nil {
		closeResources(log, db, cache)
		return nil, err
	}

	gateway := composition.BuildAWDDefenseSSHGateway(root, runtime)
	if gateway == nil {
		if runtime != nil && runtime.LifecycleCloser != nil {
			_ = runtime.LifecycleCloser.Close(ctx)
		}
		closeResources(log, db, cache)
		return nil, errors.New("awd defense ssh gateway is unavailable")
	}

	return &awdDefenseSSHGatewayProcess{
		cancel:         root.Cancel,
		gateway:        gateway,
		runtimeCloser:  runtime.LifecycleCloser,
		log:            log,
		db:             db,
		cache:          cache,
		shutdownTimout: 10 * time.Second,
	}, nil
}

func (p *awdDefenseSSHGatewayProcess) Start(ctx context.Context) error {
	if p == nil || p.gateway == nil {
		return errors.New("awd defense ssh gateway runner is required")
	}
	return p.gateway.Start(ctx)
}

func (p *awdDefenseSSHGatewayProcess) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}
	if ctx == nil {
		return errors.New("awd defense ssh gateway shutdown requires context")
	}
	var shutdownErrs []error
	drainCtx := ctx
	if p.shutdownTimout > 0 {
		var cancel context.CancelFunc
		drainCtx, cancel = context.WithTimeout(ctx, p.shutdownTimout/2)
		defer cancel()
	}
	if p.gateway != nil {
		if err := p.gateway.Drain(drainCtx); err != nil {
			shutdownErrs = append(shutdownErrs, err)
		}
	}
	if p.cancel != nil {
		p.cancel()
	}
	if p.gateway != nil {
		if err := p.gateway.Stop(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, err)
			return errors.Join(shutdownErrs...)
		}
	}
	if p.runtimeCloser != nil {
		if err := p.runtimeCloser.Close(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, err)
		}
	}
	closeResources(p.log, p.db, p.cache)
	return errors.Join(shutdownErrs...)
}

func (p *awdDefenseSSHGatewayProcess) shutdownTimeout() time.Duration {
	if p == nil || p.shutdownTimout <= 0 {
		return 10 * time.Second
	}
	return p.shutdownTimout
}
