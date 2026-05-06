package ports

import "ctf-platform/internal/model"

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
	Mounts          []model.ContainerMount
	Resources       *model.ResourceLimits
}

type TopologyCreateNetwork struct {
	Key      string
	Name     string
	Internal bool
	Shared   bool
}

type TopologyCreateRequest struct {
	Networks                   []TopologyCreateNetwork
	Nodes                      []TopologyCreateNode
	Policies                   []model.TopologyTrafficPolicy
	ReservedHostPort           int
	DisableEntryPortPublishing bool
	ContainerName              string
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}
