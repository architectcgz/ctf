package ports

import (
	"context"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

// ManagedContainerInventory 定义受管容器盘点能力。
type ManagedContainerInventory interface {
	ListManagedContainers(ctx context.Context) ([]runtimecontracts.ManagedContainer, error)
	InspectManagedContainer(ctx context.Context, containerID string) (*runtimecontracts.ManagedContainerState, error)
}
