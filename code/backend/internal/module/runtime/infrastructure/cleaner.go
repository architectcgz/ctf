package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

type Cleaner struct {
	service cleanerService
	cron    *cron.Cron
	logger  *zap.Logger
	redis   *redislib.Client
	lockTTL time.Duration
	baseCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type cleanerService interface {
	ReconcileLostActiveRuntimes(ctx context.Context) error
	CleanExpiredInstances(ctx context.Context) error
	CleanupOrphans(ctx context.Context) error
}

func NewCleaner(service cleanerService, redis *redislib.Client, lockTTL time.Duration, logger *zap.Logger) *Cleaner {
	if lockTTL <= 0 {
		lockTTL = 2 * time.Minute
	}
	return &Cleaner{
		service: service,
		cron:    cron.New(),
		logger:  logger,
		redis:   redis,
		lockTTL: lockTTL,
	}
}

func (c *Cleaner) Start(ctx context.Context, interval string) error {
	if ctx == nil {
		return errors.New("runtime cleaner start requires context")
	}
	c.baseCtx, c.cancel = context.WithCancel(ctx)
	cleanFunc := func() {
		c.startRunOnce()
	}

	_, err := c.cron.AddFunc(interval, cleanFunc)
	if err != nil {
		c.logger.Warn("cron 配置错误，使用默认间隔", zap.Error(err))
		_, err = c.cron.AddFunc("*/5 * * * *", cleanFunc)
		if err != nil {
			return fmt.Errorf("启动定时清理失败: %w", err)
		}
		interval = "*/5 * * * *"
	}
	c.cron.Start()
	c.startRunOnce()
	c.logger.Info("实例清理定时任务已启动", zap.String("interval", interval))
	return nil
}

func (c *Cleaner) startRunOnce() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.runOnce()
	}()
}

func (c *Cleaner) runOnce() {
	ctx := c.baseCtx
	if err := ctx.Err(); err != nil {
		return
	}

	lock, acquired, err := redislock.Acquire(ctx, c.redis, rediskeys.ContainerCleanupLockKey(), c.lockTTL)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			c.logger.Error("获取实例清理任务锁失败", zap.Error(err))
		}
		return
	}
	if !acquired {
		c.logger.Debug("实例清理任务已由其他节点执行")
		return
	}
	if lock != nil {
		defer func() {
			released, releaseErr := lock.Release(ctx)
			if releaseErr != nil {
				if !errors.Is(releaseErr, context.Canceled) {
					c.logger.Error("释放实例清理任务锁失败", zap.String("lock_key", lock.Key(ctx)), zap.Error(releaseErr))
				}
				return
			}
			if !released && ctx.Err() == nil {
				c.logger.Warn("实例清理任务锁已过期或被覆盖", zap.String("lock_key", lock.Key(ctx)))
			}
		}()
	}

	c.logger.Info("开始对账实例运行时")
	if err := c.service.ReconcileLostActiveRuntimes(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			c.logger.Error("对账实例运行时失败", zap.Error(err))
		}
		return
	}

	c.logger.Info("开始清理过期实例")
	if err := c.service.CleanExpiredInstances(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			c.logger.Error("清理过期实例失败", zap.Error(err))
		}
		return
	}
	if err := c.service.CleanupOrphans(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			c.logger.Error("清理孤儿容器失败", zap.Error(err))
		}
	}
}

func (c *Cleaner) Stop(ctx context.Context) error {
	if ctx == nil {
		return errors.New("runtime cleaner stop requires context")
	}
	if c.cancel != nil {
		c.cancel()
	}
	stopped := c.cron.Stop()
	select {
	case <-stopped.Done():
	case <-ctx.Done():
		return ctx.Err()
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		c.logger.Info("实例清理定时任务已停止")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
