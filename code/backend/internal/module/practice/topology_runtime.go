package practice

import (
	"fmt"
	"sort"

	"ctf-platform/internal/model"
)

type runtimeTopologyPlan struct {
	Networks        []TopologyCreateNetwork
	NodeNetworkKeys map[string][]string
}

type runtimePairPolicy struct {
	hasBroadAllow bool
	hasBroadDeny  bool
	hasFineAllow  bool
}

// buildRuntimeTopologyPlan 将逻辑拓扑中的 networks/policies 转成容器运行时网络计划。
// 当前会同时考虑粗粒度与细粒度 allow 策略来决定节点对是否需要共网，
// 再由容器层对细粒度规则补充真实的 ACL 下发。
func buildRuntimeTopologyPlan(spec model.TopologySpec) *runtimeTopologyPlan {
	logicalNetworks := normalizeLogicalTopologyNetworks(spec.Networks)
	nodeOrder := make(map[string]int, len(spec.Nodes))
	nodeNetworkMembership := make(map[string][]string, len(spec.Nodes))
	membersByNetwork := make(map[string][]string, len(logicalNetworks))
	for idx, node := range spec.Nodes {
		nodeOrder[node.Key] = idx
		networkKeys := append([]string(nil), node.NetworkKeys...)
		if len(networkKeys) == 0 {
			networkKeys = []string{logicalNetworks[0].Key}
		}
		nodeNetworkMembership[node.Key] = networkKeys
		for _, networkKey := range networkKeys {
			membersByNetwork[networkKey] = append(membersByNetwork[networkKey], node.Key)
		}
	}

	policies := indexRuntimeConnectivityPolicies(spec.Policies)
	plan := &runtimeTopologyPlan{
		Networks:        make([]TopologyCreateNetwork, 0, len(logicalNetworks)),
		NodeNetworkKeys: make(map[string][]string, len(spec.Nodes)),
	}
	addedNetworkKeys := make(map[string]struct{})

	for _, logicalNetwork := range logicalNetworks {
		members := append([]string(nil), membersByNetwork[logicalNetwork.Key]...)
		if len(members) == 0 {
			continue
		}
		sort.Slice(members, func(i, j int) bool {
			return nodeOrder[members[i]] < nodeOrder[members[j]]
		})

		explicitAllowPresent := false
		type allowedPair struct {
			left  string
			right string
		}
		allowedPairs := make([]allowedPair, 0)
		totalPairs := len(members) * (len(members) - 1) / 2
		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				rule := lookupPairPolicy(policies, members[i], members[j])
				if rule.hasBroadAllow || rule.hasFineAllow {
					explicitAllowPresent = true
				}
			}
		}

		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				rule := lookupPairPolicy(policies, members[i], members[j])
				allowed := false
				switch {
				case rule.hasBroadDeny:
					allowed = false
				case rule.hasBroadAllow || rule.hasFineAllow:
					allowed = true
				case explicitAllowPresent:
					allowed = false
				default:
					allowed = true
				}
				if allowed {
					left, right := orderedNodePair(members[i], members[j], nodeOrder)
					allowedPairs = append(allowedPairs, allowedPair{left: left, right: right})
				}
			}
		}

		if len(members) == 1 || len(allowedPairs) == totalPairs {
			attachNetwork(plan, addedNetworkKeys, logicalNetwork.Key, logicalNetwork.Internal, members...)
			continue
		}

		assigned := make(map[string]int, len(members))
		for _, pair := range allowedPairs {
			runtimeKey := fmt.Sprintf("%s__%s__%s", logicalNetwork.Key, pair.left, pair.right)
			attachNetwork(plan, addedNetworkKeys, runtimeKey, logicalNetwork.Internal, pair.left, pair.right)
			assigned[pair.left]++
			assigned[pair.right]++
		}
		for _, nodeKey := range members {
			if assigned[nodeKey] > 0 {
				continue
			}
			runtimeKey := fmt.Sprintf("%s__%s", logicalNetwork.Key, nodeKey)
			attachNetwork(plan, addedNetworkKeys, runtimeKey, logicalNetwork.Internal, nodeKey)
		}
	}

	return plan
}

func normalizeLogicalTopologyNetworks(networks []model.TopologyNetwork) []model.TopologyNetwork {
	if len(networks) > 0 {
		return networks
	}
	return []model.TopologyNetwork{
		{
			Key: model.TopologyDefaultNetworkKey,
		},
	}
}

func indexRuntimeConnectivityPolicies(policies []model.TopologyTrafficPolicy) map[string]runtimePairPolicy {
	index := make(map[string]runtimePairPolicy, len(policies))
	for _, policy := range policies {
		pairKey := unorderedNodePairKey(policy.SourceNodeKey, policy.TargetNodeKey)
		rule := index[pairKey]
		isBroad := model.IsBroadTopologyPolicy(policy)
		switch policy.Action {
		case model.TopologyPolicyActionAllow:
			if isBroad {
				rule.hasBroadAllow = true
			} else {
				rule.hasFineAllow = true
			}
		case model.TopologyPolicyActionDeny:
			if isBroad {
				rule.hasBroadDeny = true
			}
		}
		index[pairKey] = rule
	}
	return index
}

func lookupPairPolicy(index map[string]runtimePairPolicy, left, right string) runtimePairPolicy {
	return index[unorderedNodePairKey(left, right)]
}

func unorderedNodePairKey(left, right string) string {
	if left > right {
		left, right = right, left
	}
	return left + "::" + right
}

func orderedNodePair(left, right string, order map[string]int) (string, string) {
	if order[left] <= order[right] {
		return left, right
	}
	return right, left
}

func attachNetwork(plan *runtimeTopologyPlan, addedNetworkKeys map[string]struct{}, networkKey string, internal bool, nodeKeys ...string) {
	if _, exists := addedNetworkKeys[networkKey]; !exists {
		plan.Networks = append(plan.Networks, TopologyCreateNetwork{
			Key:      networkKey,
			Internal: internal,
		})
		addedNetworkKeys[networkKey] = struct{}{}
	}
	for _, nodeKey := range nodeKeys {
		plan.NodeNetworkKeys[nodeKey] = append(plan.NodeNetworkKeys[nodeKey], networkKey)
	}
}
