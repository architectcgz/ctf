package health

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
)

type Service interface {
	Check(ctx context.Context) *Status
	CheckLive(ctx context.Context) *Status
	CheckReady(ctx context.Context) *Status
	CheckDB(ctx context.Context) error
	CheckRedis(ctx context.Context) error
}

type HealthStatus struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Environment  string            `json:"environment"`
	Dependencies map[string]string `json:"dependencies"`
	Version      string            `json:"version"`
}

type DependencyCheck struct {
	Name  string
	Check func(context.Context) error
}

type Status struct {
	HealthStatus HealthStatus
	healthy      bool
}

type ReadinessState struct {
	draining atomic.Bool
}

type service struct {
	cfg       *config.Config
	db        *gorm.DB
	redis     *redislib.Client
	readiness *ReadinessState
	checks    []DependencyCheck
}

func NewReadinessState() *ReadinessState {
	return &ReadinessState{}
}

func (s *ReadinessState) MarkDraining() {
	if s == nil {
		return
	}
	s.draining.Store(true)
}

func (s *ReadinessState) IsDraining() bool {
	return s != nil && s.draining.Load()
}

func NewService(cfg *config.Config, db *gorm.DB, redis *redislib.Client, readiness *ReadinessState, checks ...DependencyCheck) Service {
	if readiness == nil {
		readiness = NewReadinessState()
	}
	filteredChecks := make([]DependencyCheck, 0, len(checks))
	for _, check := range checks {
		if check.Name == "" || check.Check == nil {
			continue
		}
		filteredChecks = append(filteredChecks, check)
	}
	return &service{
		cfg:       cfg,
		db:        db,
		redis:     redis,
		readiness: readiness,
		checks:    filteredChecks,
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
	for _, check := range s.checks {
		dependencies[check.Name] = "ok"
		if err := check.Check(ctx); err != nil {
			dependencies[check.Name] = "down"
			healthy = false
		}
	}

	status := "ok"
	if !healthy {
		status = "degraded"
	}

	return NewStatus(s.newHealthStatus(status, dependencies), healthy)
}

func (s *service) CheckLive(context.Context) *Status {
	return NewStatus(s.newHealthStatus("ok", map[string]string{
		"process": "ok",
	}), true)
}

func (s *service) CheckReady(ctx context.Context) *Status {
	dependencies := map[string]string{
		"process":  "ok",
		"postgres": "ok",
		"redis":    "ok",
	}
	ready := true

	if s.readiness.IsDraining() {
		dependencies["process"] = "draining"
		ready = false
	}
	if err := s.CheckDB(ctx); err != nil {
		dependencies["postgres"] = "down"
		ready = false
	}
	if err := s.CheckRedis(ctx); err != nil {
		dependencies["redis"] = "down"
		ready = false
	}
	for _, check := range s.checks {
		dependencies[check.Name] = "ok"
		if err := check.Check(ctx); err != nil {
			dependencies[check.Name] = "down"
			ready = false
		}
	}

	status := "ready"
	if !ready {
		status = "not_ready"
	}

	return NewStatus(s.newHealthStatus(status, dependencies), ready)
}

func (s *service) CheckDB(ctx context.Context) error {
	if s.db == nil {
		return errors.New("postgres dependency is not configured")
	}
	sqlDB, err := s.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (s *service) CheckRedis(ctx context.Context) error {
	if s.redis == nil {
		return errors.New("redis dependency is not configured")
	}
	return s.redis.Ping(ctx).Err()
}

func (s *service) newHealthStatus(status string, dependencies map[string]string) HealthStatus {
	return HealthStatus{
		Status:       status,
		Service:      s.cfg.App.Name,
		Environment:  s.cfg.App.Env,
		Dependencies: dependencies,
		Version:      s.cfg.App.Version,
	}
}

func NewStatus(status HealthStatus, healthy bool) *Status {
	return &Status{
		HealthStatus: status,
		healthy:      healthy,
	}
}

func (s *Status) HTTPStatus() int {
	if s.healthy {
		return http.StatusOK
	}
	return http.StatusServiceUnavailable
}
