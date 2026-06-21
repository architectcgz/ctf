package contracts

// RuntimeCleanupTarget is the runtime-owned view needed to clean containers,
// networks, ACL state, and reserved allocations for one instance runtime.
type RuntimeCleanupTarget struct {
	InstanceID     int64
	RuntimeNodeID  *int64
	ContainerID    string
	NetworkID      string
	HostPort       int
	RuntimeDetails string
}
