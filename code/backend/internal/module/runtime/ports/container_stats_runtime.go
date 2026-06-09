package ports

import "context"

// ManagedContainerStatsReader 定义受管容器指标读取能力。
type ManagedContainerStatsReader interface {
	ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error)
}
