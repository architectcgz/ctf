package ports

import (
	"context"

	"ctf-platform/internal/config"
)

type RuntimeResourcePoolRepository interface {
	EnsurePoolsForNode(ctx context.Context, nodeID int64, cfg config.ContainerConfig) error
	ReserveAvailablePortForNode(ctx context.Context, nodeID, instanceID int64) (int, error)
	ReserveAvailableSubnetForNode(ctx context.Context, nodeID int64, poolKind string, instanceID int64, networkKey string) (string, error)
	BindResourcesForInstance(ctx context.Context, instanceID int64) error
	ReleaseResourcesForInstance(ctx context.Context, instanceID int64) error
	QuarantinePort(ctx context.Context, nodeID int64, port int, reason string) error
	QuarantineSubnet(ctx context.Context, nodeID int64, subnet string, reason string) error
}
