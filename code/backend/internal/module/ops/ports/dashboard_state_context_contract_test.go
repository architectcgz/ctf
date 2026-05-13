package ports_test

import (
	"context"

	opsports "ctf-platform/internal/module/ops/ports"
)

type ctxOnlyDashboardStateStore struct{}

func (ctxOnlyDashboardStateStore) LoadDashboardStats(context.Context) (*opsports.DashboardStatsSnapshot, error) {
	return nil, nil
}

func (ctxOnlyDashboardStateStore) SaveDashboardStats(context.Context, *opsports.DashboardStatsSnapshot) error {
	return nil
}

func (ctxOnlyDashboardStateStore) CountOnlineUsers(context.Context) (int64, error) {
	return 0, nil
}

var _ opsports.DashboardStateStore = (*ctxOnlyDashboardStateStore)(nil)
