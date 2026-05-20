package ports

import runtimecontracts "ctf-platform/internal/module/runtime/contracts"

type TopologyCreateNode struct {
	Key             string
	Image           string
	Env             map[string]string
	Command         []string
	WorkingDir      string
	ServicePort     int
	ServiceProtocol string
	IsEntryPoint    bool
	NetworkKeys     []string
	NetworkAliases  []string
	Mounts          []runtimecontracts.ContainerMount
	Resources       *runtimecontracts.ResourceLimits
}

type TopologyCreateNetwork struct {
	Key      string
	Name     string
	Subnet   string
	Internal bool
	Shared   bool
}

type TopologyCreateRequest struct {
	Networks                   []TopologyCreateNetwork
	Nodes                      []TopologyCreateNode
	Policies                   []runtimecontracts.TopologyTrafficPolicy
	OwnerInstanceID            int64
	ReservedHostPort           int
	DisableEntryPortPublishing bool
	ContainerName              string
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     runtimecontracts.InstanceRuntimeDetails
}
