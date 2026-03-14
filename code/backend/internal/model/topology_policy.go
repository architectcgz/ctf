package model

import "strings"

// IsBroadTopologyPolicy reports whether a policy can be enforced by the current
// runtime network segmentation implementation.
func IsBroadTopologyPolicy(policy TopologyTrafficPolicy) bool {
	if len(policy.Ports) > 0 {
		return false
	}
	protocol := strings.TrimSpace(policy.Protocol)
	return protocol == "" || protocol == TopologyPolicyProtocolAny
}

// FirstFineGrainedTopologyPolicy returns the first policy that requires
// protocol/port-level ACL enforcement, which is not yet supported at runtime.
func FirstFineGrainedTopologyPolicy(spec TopologySpec) *TopologyTrafficPolicy {
	for idx := range spec.Policies {
		if IsBroadTopologyPolicy(spec.Policies[idx]) {
			continue
		}
		policy := spec.Policies[idx]
		return &policy
	}
	return nil
}
