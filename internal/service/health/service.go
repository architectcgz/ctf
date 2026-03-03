package health

import (
	"context"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

type Service interface {
	Check(ctx context.Context) *model.HealthStatus
}

type service struct {
	cfg *config.Config
}

func NewService(cfg *config.Config) Service {
	return &service{cfg: cfg}
}

func (s *service) Check(_ context.Context) *model.HealthStatus {
	return &model.HealthStatus{
		Status:      "ok",
		Service:     s.cfg.App.Name,
		Environment: s.cfg.App.Env,
		Version:     "dev",
		Dependencies: map[string]string{
			"postgres": "configured",
			"redis":    "configured",
		},
	}
}
