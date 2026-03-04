package health

import (
	"context"
	"net/http"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

type Service interface {
	Check(ctx context.Context) *Status
	CheckDB(ctx context.Context) error
	CheckRedis(ctx context.Context) error
}

type Status struct {
	HealthStatus model.HealthStatus
	healthy      bool
}

type service struct {
	cfg   *config.Config
	db    *gorm.DB
	redis *redislib.Client
}

func NewService(cfg *config.Config, db *gorm.DB, redis *redislib.Client) Service {
	return &service{
		cfg:   cfg,
		db:    db,
		redis: redis,
	}
}

func (s *service) Check(ctx context.Context) *Status {
	dependencies := map[string]string{
		"postgres": "ok",
		"redis":    "ok",
	}
	healthy := true

	if err := s.CheckDB(ctx); err != nil {
		dependencies["postgres"] = "down"
		healthy = false
	}
	if err := s.CheckRedis(ctx); err != nil {
		dependencies["redis"] = "down"
		healthy = false
	}

	status := "ok"
	if !healthy {
		status = "degraded"
	}

	return &Status{
		HealthStatus: model.HealthStatus{
			Status:       status,
			Service:      s.cfg.App.Name,
			Environment:  s.cfg.App.Env,
			Dependencies: dependencies,
			Version:      s.cfg.App.Version,
		},
		healthy: healthy,
	}
}

func (s *service) CheckDB(ctx context.Context) error {
	sqlDB, err := s.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (s *service) CheckRedis(ctx context.Context) error {
	return s.redis.Ping(ctx).Err()
}

func (s *Status) HTTPStatus() int {
	if s.healthy {
		return http.StatusOK
	}
	return http.StatusServiceUnavailable
}
