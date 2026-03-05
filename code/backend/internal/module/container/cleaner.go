package container

import (
	"context"

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

func (c *Cleaner) Start() error {
	_, err := c.cron.AddFunc("*/5 * * * *", func() {
		c.logger.Info("开始清理过期实例")
		if err := c.service.CleanExpiredInstances(context.Background()); err != nil {
			c.logger.Error("清理过期实例失败", zap.Error(err))
		}
	})
	if err != nil {
		return err
	}
	c.cron.Start()
	c.logger.Info("实例清理定时任务已启动")
	return nil
}

func (c *Cleaner) Stop() {
	c.cron.Stop()
	c.logger.Info("实例清理定时任务已停止")
}
