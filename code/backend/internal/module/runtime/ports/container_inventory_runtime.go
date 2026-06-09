package ports

import "context"

// ManagedContainerInventory 定义受管容器盘点能力。
type ManagedContainerInventory interface {
	ListManagedContainers(ctx context.Context) ([]ManagedContainer, error)
	InspectManagedContainer(ctx context.Context, containerID string) (*ManagedContainerState, error)
}
