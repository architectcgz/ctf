package commands

import (
	"context"
	"fmt"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"ctf-platform/pkg/errcode"
)

// RuntimeMaintenanceService 保留 runtime compat import path，并把实例维护能力委托给 instance owner。
type RuntimeMaintenanceService struct {
	service instancecontracts.MaintenanceService
}

func NewRuntimeMaintenanceService(service instancecontracts.MaintenanceService) *RuntimeMaintenanceService {
	return &RuntimeMaintenanceService{service: service}
}

func (s *RuntimeMaintenanceService) CleanExpiredInstances(ctx context.Context) error {
	if s == nil || s.service == nil {
		return errRuntimeCompatMaintenanceServiceUnavailable()
	}
	return s.service.CleanExpiredInstances(ctx)
}

func (s *RuntimeMaintenanceService) ReconcileLostActiveRuntimes(ctx context.Context) error {
	if s == nil || s.service == nil {
		return errRuntimeCompatMaintenanceServiceUnavailable()
	}
	return s.service.ReconcileLostActiveRuntimes(ctx)
}

func (s *RuntimeMaintenanceService) CleanupOrphans(ctx context.Context) error {
	if s == nil || s.service == nil {
		return errRuntimeCompatMaintenanceServiceUnavailable()
	}
	return s.service.CleanupOrphans(ctx)
}

func errRuntimeCompatMaintenanceServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("runtime maintenance compat service is not configured"))
}
