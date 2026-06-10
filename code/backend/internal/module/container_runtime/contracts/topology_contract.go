package contracts

import "strings"

const (
	ChallengeTargetProtocolHTTP = "http"
	ChallengeTargetProtocolTCP  = "tcp"

	TopologyDefaultNetworkKey = "default"

	TopologyPolicyActionAllow = "allow"
	TopologyPolicyActionDeny  = "deny"

	TopologyPolicyProtocolTCP = "tcp"
	TopologyPolicyProtocolUDP = "udp"
	TopologyPolicyProtocolAny = "any"
)

type TopologyTrafficPolicy struct {
	SourceNodeKey string `json:"source_node_key"`
	TargetNodeKey string `json:"target_node_key"`
	Action        string `json:"action"`
	Protocol      string `json:"protocol,omitempty"`
	Ports         []int  `json:"ports,omitempty"`
}

func IsBroadTopologyPolicy(policy TopologyTrafficPolicy) bool {
	if len(policy.Ports) > 0 {
		return false
	}
	protocol := strings.TrimSpace(policy.Protocol)
	return protocol == "" || protocol == TopologyPolicyProtocolAny
}
