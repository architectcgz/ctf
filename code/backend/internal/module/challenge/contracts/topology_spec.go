package contracts

import (
	"encoding/json"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
)

const (
	TopologyTierPublic   = "public"
	TopologyTierService  = "service"
	TopologyTierInternal = "internal"

	TopologyDefaultNetworkKey = runtimecontracts.TopologyDefaultNetworkKey

	TopologyPolicyActionAllow = runtimecontracts.TopologyPolicyActionAllow
	TopologyPolicyActionDeny  = runtimecontracts.TopologyPolicyActionDeny

	TopologyPolicyProtocolTCP = runtimecontracts.TopologyPolicyProtocolTCP
	TopologyPolicyProtocolUDP = runtimecontracts.TopologyPolicyProtocolUDP
	TopologyPolicyProtocolAny = runtimecontracts.TopologyPolicyProtocolAny
)

type TopologySpec struct {
	Networks []TopologyNetwork       `json:"networks,omitempty"`
	Nodes    []TopologyNode          `json:"nodes"`
	Links    []TopologyLink          `json:"links,omitempty"`
	Policies []TopologyTrafficPolicy `json:"policies,omitempty"`
}

type TopologyNetwork struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	CIDR     string `json:"cidr,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}

type TopologyNode struct {
	Key             string             `json:"key"`
	Name            string             `json:"name"`
	ImageID         int64              `json:"image_id,omitempty"`
	ServicePort     int                `json:"service_port,omitempty"`
	ServiceProtocol string             `json:"service_protocol,omitempty"`
	InjectFlag      bool               `json:"inject_flag,omitempty"`
	Tier            string             `json:"tier,omitempty"`
	NetworkKeys     []string           `json:"network_keys,omitempty"`
	Env             map[string]string  `json:"env,omitempty"`
	Resources       *TopologyResources `json:"resources,omitempty"`
}

type TopologyResources struct {
	CPUQuota  float64 `json:"cpu_quota,omitempty"`
	MemoryMB  int64   `json:"memory_mb,omitempty"`
	PidsLimit int64   `json:"pids_limit,omitempty"`
}

type TopologyLink struct {
	FromNodeKey string `json:"from_node_key"`
	ToNodeKey   string `json:"to_node_key"`
}

type TopologyTrafficPolicy = runtimecontracts.TopologyTrafficPolicy

func EncodeTopologySpec(spec TopologySpec) (string, error) {
	raw, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func DecodeTopologySpec(raw string) (TopologySpec, error) {
	if raw == "" {
		return TopologySpec{}, nil
	}
	var spec TopologySpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return TopologySpec{}, err
	}
	return spec, nil
}

func IsBroadTopologyPolicy(policy TopologyTrafficPolicy) bool {
	return runtimecontracts.IsBroadTopologyPolicy(policy)
}
