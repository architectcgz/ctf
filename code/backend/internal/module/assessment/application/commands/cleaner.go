package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Cleaner struct {
	service cleanerService
	cron    *cron.Cron
	logger  *zap.Logger
	baseCtx context.Context
	cancel  context.CancelFunc
}

type cleanerService interface {
	RebuildAllSkillProfiles(ctx context.Context) error
}

func NewCleaner(service cleanerService, logger *zap.Logger) *Cleaner {
	if logger == nil {
		logger = zap.NewNop()
	}
	baseCtx, cancel := context.WithCancel(context.Background())
	return &Cleaner{
		service: service,
		cron:    cron.New(),
		logger:  logger,
		baseCtx: baseCtx,
		cancel:  cancel,
	}
}

func (c *Cleaner) Start(spec string, timeout time.Duration) error {
	rebuild := func() {
		c.runOnce(timeout)
	}

	_, err := c.cron.AddFunc(spec, rebuild)
	if err != nil {
		c.logger.Warn("能力画像 cron 配置错误，使用默认间隔", zap.Error(err))
		spec = "0 0 * * *"
		_, err = c.cron.AddFunc(spec, rebuild)
		if err != nil {
			return fmt.Errorf("启动能力画像定时任务失败: %w", err)
		}
	}

	c.cron.Start()
	c.logger.Info("能力画像定时任务已启动", zap.String("cron", spec), zap.Duration("timeout", timeout))
	return nil
}

func (c *Cleaner) runOnce(timeout time.Duration) {
	if err := c.baseCtx.Err(); err != nil {
		return
	}

	ctx := c.baseCtx
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(c.baseCtx, timeout)
	}
	defer cancel()

	c.logger.Info("开始重建能力画像")
	if err := c.service.RebuildAllSkillProfiles(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			c.logger.Error("重建能力画像失败", zap.Error(err))
		}
		return
	}
	c.logger.Info("能力画像重建完成")
}

func (c *Cleaner) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.cancel()
	stopped := c.cron.Stop()
	select {
	case <-stopped.Done():
		c.logger.Info("能力画像定时任务已停止")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
