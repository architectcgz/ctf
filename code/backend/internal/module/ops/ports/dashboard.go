package ports

import "context"

type ManagedContainerStat struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   int64
	MemoryLimit   int64
}

type RuntimeQuery interface {
	CountRunning(ctx context.Context) (int64, error)
}

type RuntimeStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error)
}

type DashboardContainerStat struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   int64
	MemoryLimit   int64
}

type DashboardResourceAlert struct {
	ContainerID string
	Type        string
	Value       float64
	Threshold   float64
	Message     string
}

type DashboardStatsSnapshot struct {
	OnlineUsers      int64
	ActiveContainers int64
	CPUUsage         float64
	MemoryUsage      float64
	ContainerStats   []DashboardContainerStat
	Alerts           []DashboardResourceAlert
}

type DashboardStateStore interface {
	LoadDashboardStats(ctx context.Context) (*DashboardStatsSnapshot, error)
	SaveDashboardStats(ctx context.Context, stats *DashboardStatsSnapshot) error
	CountOnlineUsers(ctx context.Context) (int64, error)
}
