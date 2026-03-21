package runtime

import (
	"context"

	"ctf-platform/internal/module/container"
)

type ManagedContainerStat = container.ManagedContainerStat

type RuntimeStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error)
}
