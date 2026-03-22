package domain

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"ctf-platform/internal/model"
)

type aclRuleGroup struct {
	sourceNodeKey     string
	targetNodeKey     string
	sourceContainerID string
	targetContainerID string
	sourceIP          string
	targetIP          string
	denyRules         []aclRuleSpec
	allowRules        []aclRuleSpec
}

type aclRuleSpec struct {
	protocol string
	ports    []int
}

type aclBinding struct {
	sourceIP string
	targetIP string
}

// ResolveTopologyACLRules 根据实例运行时拓扑和容器 IP 信息解析细粒度 ACL 规则。
func ResolveTopologyACLRules(
	policies []model.TopologyTrafficPolicy,
	details model.InstanceRuntimeDetails,
	ipsByContainerID map[string]map[string]string,
) ([]model.InstanceRuntimeACLRule, error) {
	if len(policies) == 0 {
		return nil, nil
	}

	containerByNodeKey := make(map[string]model.InstanceRuntimeContainer, len(details.Containers))
	for _, container := range details.Containers {
		containerByNodeKey[container.NodeKey] = container
	}

	networkNameByKey := make(map[string]string, len(details.Networks))
	for _, network := range details.Networks {
		if strings.TrimSpace(network.Key) == "" || strings.TrimSpace(network.Name) == "" {
			continue
		}
		networkNameByKey[network.Key] = network.Name
	}

	groups := make(map[string]*aclRuleGroup)
	for _, policy := range policies {
		if model.IsBroadTopologyPolicy(policy) {
			continue
		}

		source, sourceExists := containerByNodeKey[policy.SourceNodeKey]
		target, targetExists := containerByNodeKey[policy.TargetNodeKey]
		if !sourceExists || !targetExists {
			return nil, fmt.Errorf("topology acl references missing runtime container: %s -> %s", policy.SourceNodeKey, policy.TargetNodeKey)
		}

		sharedBindings := sharedRuntimeACLBindings(source, target, networkNameByKey, ipsByContainerID)
		if len(sharedBindings) == 0 {
			if policy.Action == model.TopologyPolicyActionAllow {
				return nil, fmt.Errorf("topology acl has no shared runtime network: %s -> %s", policy.SourceNodeKey, policy.TargetNodeKey)
			}
			continue
		}

		for _, binding := range sharedBindings {
			groupKey := binding.groupKey()
			group := groups[groupKey]
			if group == nil {
				group = &aclRuleGroup{
					sourceNodeKey:     source.NodeKey,
					targetNodeKey:     target.NodeKey,
					sourceContainerID: source.ContainerID,
					targetContainerID: target.ContainerID,
					sourceIP:          binding.sourceIP,
					targetIP:          binding.targetIP,
				}
				groups[groupKey] = group
			}

			for _, protocol := range expandACLProtocols(policy.Protocol, policy.Ports) {
				spec := aclRuleSpec{
					protocol: protocol,
					ports:    append([]int(nil), policy.Ports...),
				}
				switch policy.Action {
				case model.TopologyPolicyActionAllow:
					group.allowRules = appendUniqueACLRule(group.allowRules, spec)
				case model.TopologyPolicyActionDeny:
					group.denyRules = appendUniqueACLRule(group.denyRules, spec)
				default:
					return nil, fmt.Errorf("unsupported topology acl action: %s", policy.Action)
				}
			}
		}
	}

	if len(groups) == 0 {
		return nil, nil
	}

	groupKeys := make([]string, 0, len(groups))
	for key := range groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	rules := make([]model.InstanceRuntimeACLRule, 0)
	for _, key := range groupKeys {
		group := groups[key]
		for _, spec := range group.denyRules {
			rules = append(rules, newRuntimeACLRule(group, model.TopologyPolicyActionDeny, spec.protocol, spec.ports))
		}
		for _, spec := range group.allowRules {
			rules = append(rules, newRuntimeACLRule(group, model.TopologyPolicyActionAllow, spec.protocol, spec.ports))
		}
		if len(group.allowRules) > 0 {
			rules = append(rules, newRuntimeACLRule(group, model.TopologyPolicyActionDeny, model.TopologyPolicyProtocolAny, nil))
		}
	}

	return rules, nil
}

func (b aclBinding) groupKey() string {
	return b.sourceIP + "->" + b.targetIP
}

func sharedRuntimeACLBindings(
	source model.InstanceRuntimeContainer,
	target model.InstanceRuntimeContainer,
	networkNameByKey map[string]string,
	ipsByContainerID map[string]map[string]string,
) []aclBinding {
	if source.ContainerID == "" || target.ContainerID == "" {
		return nil
	}

	targetNetworkKeys := make(map[string]struct{}, len(target.NetworkKeys))
	for _, networkKey := range target.NetworkKeys {
		targetNetworkKeys[networkKey] = struct{}{}
	}

	sourceIPs := ipsByContainerID[source.ContainerID]
	targetIPs := ipsByContainerID[target.ContainerID]
	bindings := make([]aclBinding, 0)
	seen := make(map[string]struct{})
	for _, networkKey := range source.NetworkKeys {
		if _, exists := targetNetworkKeys[networkKey]; !exists {
			continue
		}
		networkName := networkNameByKey[networkKey]
		sourceIP := strings.TrimSpace(sourceIPs[networkName])
		targetIP := strings.TrimSpace(targetIPs[networkName])
		if sourceIP == "" || targetIP == "" {
			continue
		}
		binding := aclBinding{sourceIP: sourceIP, targetIP: targetIP}
		key := binding.groupKey()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		bindings = append(bindings, binding)
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].sourceIP == bindings[j].sourceIP {
			return bindings[i].targetIP < bindings[j].targetIP
		}
		return bindings[i].sourceIP < bindings[j].sourceIP
	})
	return bindings
}

func expandACLProtocols(protocol string, ports []int) []string {
	normalized := strings.TrimSpace(protocol)
	if normalized == "" {
		normalized = model.TopologyPolicyProtocolAny
	}
	if normalized == model.TopologyPolicyProtocolAny && len(ports) > 0 {
		return []string{model.TopologyPolicyProtocolTCP, model.TopologyPolicyProtocolUDP}
	}
	return []string{normalized}
}

func appendUniqueACLRule(items []aclRuleSpec, candidate aclRuleSpec) []aclRuleSpec {
	for _, item := range items {
		if item.protocol != candidate.protocol {
			continue
		}
		if sameACLPorts(item.ports, candidate.ports) {
			return items
		}
	}
	return append(items, candidate)
}

func sameACLPorts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}

func newRuntimeACLRule(group *aclRuleGroup, action, protocol string, ports []int) model.InstanceRuntimeACLRule {
	rule := model.InstanceRuntimeACLRule{
		SourceNodeKey:     group.sourceNodeKey,
		TargetNodeKey:     group.targetNodeKey,
		SourceContainerID: group.sourceContainerID,
		TargetContainerID: group.targetContainerID,
		SourceIP:          group.sourceIP,
		TargetIP:          group.targetIP,
		Action:            action,
		Protocol:          protocol,
		Ports:             append([]int(nil), ports...),
	}
	rule.Comment = buildRuntimeACLComment(rule)
	return rule
}

func buildRuntimeACLComment(rule model.InstanceRuntimeACLRule) string {
	payload := strings.Join([]string{
		rule.SourceContainerID,
		rule.TargetContainerID,
		rule.SourceIP,
		rule.TargetIP,
		rule.Action,
		rule.Protocol,
		joinACLPorts(rule.Ports),
	}, "|")
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(payload))
	return "ctf:acl:" + strconv.FormatUint(hasher.Sum64(), 16)
}

func joinACLPorts(ports []int) string {
	if len(ports) == 0 {
		return ""
	}
	items := make([]string, 0, len(ports))
	for _, port := range ports {
		items = append(items, strconv.Itoa(port))
	}
	return strings.Join(items, ",")
}
