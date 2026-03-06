package container

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Cleaner struct {
	service *Service
	cron    *cron.Cron
	logger  *zap.Logger
}

func NewCleaner(service *Service, logger *zap.Logger) *Cleaner {
	return &Cleaner{
		service: service,
		cron:    cron.New(),
		logger:  logger,
	}
}

func (c *Cleaner) Start(interval string) error {
	cleanFunc := func() {
		c.logger.Info("开始清理过期实例")
		if err := c.service.CleanExpiredInstances(context.Background()); err != nil {
			c.logger.Error("清理过期实例失败", zap.Error(err))
		}
		if err := c.service.CleanupOrphans(context.Background()); err != nil {
			c.logger.Error("清理孤儿容器失败", zap.Error(err))
		}
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
	c.logger.Info("实例清理定时任务已启动", zap.String("interval", interval))
	return nil
}

func (c *Cleaner) Stop() {
	c.cron.Stop()
	c.logger.Info("实例清理定时任务已停止")
}
