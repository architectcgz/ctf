package assessment

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Cleaner struct {
	service *Service
	cron    *cron.Cron
	logger  *zap.Logger
}

func NewCleaner(service *Service, logger *zap.Logger) *Cleaner {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Cleaner{
		service: service,
		cron:    cron.New(),
		logger:  logger,
	}
}

func (c *Cleaner) Start(spec string, timeout time.Duration) error {
	rebuild := func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		c.logger.Info("开始重建能力画像")
		if err := c.service.RebuildAllSkillProfiles(ctx); err != nil {
			c.logger.Error("重建能力画像失败", zap.Error(err))
			return
		}
		c.logger.Info("能力画像重建完成")
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

func (c *Cleaner) Stop() {
	c.cron.Stop()
	c.logger.Info("能力画像定时任务已停止")
}
