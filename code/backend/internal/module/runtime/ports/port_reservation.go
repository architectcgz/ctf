package ports

import "context"

type PortReservationOwner interface {
	ReserveAvailablePort(ctx context.Context, start, end int) (int, error)
	ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error)
	BindReservedPort(ctx context.Context, port int, instanceID int64) error
	ReleaseReservedPort(ctx context.Context, port int) error
	ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error
	IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error)
	SyncInstanceHostPortForRestart(ctx context.Context, instanceID int64, hostPort int, preserveHostPort bool) (int, error)
}

type NetworkReservationOwner interface {
	ReserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int) (string, error)
	ReserveAvailableSubnetForInstance(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string) (string, error)
	ReleaseReservedSubnet(ctx context.Context, subnet string) error
	ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error
}
