package practice

import "ctf-platform/internal/model"

type TopologyCreateNode struct {
	Key          string
	Image        string
	Env          map[string]string
	ServicePort  int
	IsEntryPoint bool
	NetworkKeys  []string
	Resources    *model.ResourceLimits
}

type TopologyCreateNetwork struct {
	Key      string
	Internal bool
}

type TopologyCreateRequest struct {
	Networks         []TopologyCreateNetwork
	Nodes            []TopologyCreateNode
	Policies         []model.TopologyTrafficPolicy
	ReservedHostPort int
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}
