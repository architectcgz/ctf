package ports

import (
	"context"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

// ManagedContainerStatsReader 定义受管容器指标读取能力。
type ManagedContainerStatsReader interface {
	ListManagedContainerStats(ctx context.Context) ([]runtimecontracts.ManagedContainerStat, error)
}
