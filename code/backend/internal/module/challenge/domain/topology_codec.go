package domain

import (
	"errors"
	"strings"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

func BuildTopologySpec(entryNodeKey string, networks []challengecontracts.TopologyNetworkReq, nodes []challengecontracts.TopologyNodeReq, links []challengecontracts.TopologyLinkReq, policies []challengecontracts.TopologyTrafficPolicyReq) (string, string, error) {
	spec, normalizedEntryNodeKey, err := normalizeTopologySpec(entryNodeKey, networks, nodes, links, policies)
	if err != nil {
		return "", "", err
	}
	raw, err := challengecontracts.EncodeTopologySpec(spec)
	if err != nil {
		return "", "", err
	}
	return raw, normalizedEntryNodeKey, nil
}

func normalizeTopologySpec(entryNodeKey string, networks []challengecontracts.TopologyNetworkReq, nodes []challengecontracts.TopologyNodeReq, links []challengecontracts.TopologyLinkReq, policies []challengecontracts.TopologyTrafficPolicyReq) (challengecontracts.TopologySpec, string, error) {
	if len(nodes) == 0 {
		return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("拓扑至少需要一个节点"))
	}

	specNetworks, networkKeys, fallbackNetworkKey, err := normalizeTopologyNetworks(networks)
	if err != nil {
		return challengecontracts.TopologySpec{}, "", err
	}

	seenNodes := make(map[string]struct{}, len(nodes))
	specNodes := make([]challengecontracts.TopologyNode, 0, len(nodes))
	injectFlagCount := 0
	for _, node := range nodes {
		key := strings.TrimSpace(node.Key)
		if key == "" {
			return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("节点 key 不能为空"))
		}
		if _, exists := seenNodes[key]; exists {
			return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("节点 key 不能重复"))
		}
		seenNodes[key] = struct{}{}

		tier := strings.TrimSpace(node.Tier)
		if tier == "" {
			tier = challengecontracts.TopologyTierService
		}
		networkRefs, networkErr := normalizeNodeNetworks(node.NetworkKeys, networkKeys, fallbackNetworkKey)
		if networkErr != nil {
			return challengecontracts.TopologySpec{}, "", networkErr
		}

		var resources *challengecontracts.TopologyResources
		if node.Resources != nil {
			resources = &challengecontracts.TopologyResources{
				CPUQuota:  node.Resources.CPUQuota,
				MemoryMB:  node.Resources.MemoryMB,
				PidsLimit: node.Resources.PidsLimit,
			}
		}
		if node.InjectFlag {
			injectFlagCount++
		}

		specNodes = append(specNodes, challengecontracts.TopologyNode{
			Key:             key,
			Name:            strings.TrimSpace(node.Name),
			ImageID:         node.ImageID,
			ServicePort:     node.ServicePort,
			ServiceProtocol: normalizeServiceProtocol(node.ServiceProtocol),
			InjectFlag:      node.InjectFlag,
			Tier:            tier,
			NetworkKeys:     networkRefs,
			Env:             trimEnvMap(node.Env),
			Resources:       resources,
		})
	}

	entry := strings.TrimSpace(entryNodeKey)
	if entry == "" {
		entry = specNodes[0].Key
	}
	if _, exists := seenNodes[entry]; !exists {
		return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("入口节点不存在"))
	}

	if injectFlagCount == 0 {
		for idx := range specNodes {
			if specNodes[idx].Key == entry {
				specNodes[idx].InjectFlag = true
				break
			}
		}
	}

	specLinks := make([]challengecontracts.TopologyLink, 0, len(links))
	for _, link := range links {
		from := strings.TrimSpace(link.FromNodeKey)
		to := strings.TrimSpace(link.ToNodeKey)
		if _, exists := seenNodes[from]; !exists {
			return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("拓扑连线源节点不存在"))
		}
		if _, exists := seenNodes[to]; !exists {
			return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("拓扑连线目标节点不存在"))
		}
		specLinks = append(specLinks, challengecontracts.TopologyLink{
			FromNodeKey: from,
			ToNodeKey:   to,
		})
	}

	specPolicies := make([]challengecontracts.TopologyTrafficPolicy, 0, len(policies))
	for _, policy := range policies {
		source := strings.TrimSpace(policy.SourceNodeKey)
		target := strings.TrimSpace(policy.TargetNodeKey)
		if _, exists := seenNodes[source]; !exists {
			return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("链路策略源节点不存在"))
		}
		if _, exists := seenNodes[target]; !exists {
			return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("链路策略目标节点不存在"))
		}
		specPolicies = append(specPolicies, challengecontracts.TopologyTrafficPolicy{
			SourceNodeKey: source,
			TargetNodeKey: target,
			Action:        strings.TrimSpace(policy.Action),
			Protocol:      normalizePolicyProtocol(policy.Protocol),
			Ports:         normalizePolicyPorts(policy.Ports),
		})
	}
	hasEntryPort := false
	for _, node := range specNodes {
		if node.Key == entry && node.ServicePort > 0 {
			hasEntryPort = true
			break
		}
	}
	if !hasEntryPort {
		return challengecontracts.TopologySpec{}, "", apperror.ErrInvalidParams.WithCause(errors.New("入口节点必须配置 service_port"))
	}

	return challengecontracts.TopologySpec{
		Networks: specNetworks,
		Nodes:    specNodes,
		Links:    specLinks,
		Policies: specPolicies,
	}, entry, nil
}

func TopologyRespFromModel(item *challengeentity.ChallengeTopology) (*challengecontracts.ChallengeTopologyResp, error) {
	spec, err := challengecontracts.DecodeTopologySpec(item.Spec)
	if err != nil {
		return nil, err
	}

	resp := challengeResponseMapperInst.ToChallengeTopologyRespBasePtr(item)
	var baseline *challengecontracts.TopologySpecResp
	if strings.TrimSpace(item.PackageBaselineSpec) != "" {
		baselineSpec, decodeErr := challengecontracts.DecodeTopologySpec(item.PackageBaselineSpec)
		if decodeErr != nil {
			return nil, decodeErr
		}
		baseline = topologySpecRespFromSpec(item.EntryNodeKey, baselineSpec)
	}
	resp.Networks = topologyNetworkRespList(spec.Networks)
	resp.Nodes = topologyNodeRespList(spec.Nodes)
	resp.Links = topologyLinkRespList(spec.Links)
	resp.Policies = topologyTrafficPolicyRespList(spec.Policies)
	resp.PackageBaseline = baseline
	return resp, nil
}

func TopologySpecRespFromEncoded(entryNodeKey string, raw string) (*challengecontracts.TopologySpecResp, error) {
	spec, err := challengecontracts.DecodeTopologySpec(raw)
	if err != nil {
		return nil, err
	}
	return topologySpecRespFromSpec(entryNodeKey, spec), nil
}

func TopologySpecRespFromSpec(entryNodeKey string, spec challengecontracts.TopologySpec) *challengecontracts.TopologySpecResp {
	return topologySpecRespFromSpec(entryNodeKey, spec)
}

func TemplateRespFromModel(item *challengeentity.EnvironmentTemplate) (*challengecontracts.EnvironmentTemplateResp, error) {
	spec, err := challengecontracts.DecodeTopologySpec(item.Spec)
	if err != nil {
		return nil, err
	}
	resp := challengeResponseMapperInst.ToEnvironmentTemplateRespBasePtr(item)
	resp.Networks = topologyNetworkRespList(spec.Networks)
	resp.Nodes = topologyNodeRespList(spec.Nodes)
	resp.Links = topologyLinkRespList(spec.Links)
	resp.Policies = topologyTrafficPolicyRespList(spec.Policies)
	return resp, nil
}

func topologyNodeRespList(nodes []challengecontracts.TopologyNode) []challengecontracts.TopologyNodeResp {
	return challengeResponseMapperInst.ToTopologyNodeResps(nodes)
}

func topologyNetworkRespList(networks []challengecontracts.TopologyNetwork) []challengecontracts.TopologyNetworkResp {
	return challengeResponseMapperInst.ToTopologyNetworkResps(networks)
}

func topologyLinkRespList(links []challengecontracts.TopologyLink) []challengecontracts.TopologyLinkResp {
	return challengeResponseMapperInst.ToTopologyLinkResps(links)
}

func topologyTrafficPolicyRespList(policies []challengecontracts.TopologyTrafficPolicy) []challengecontracts.TopologyTrafficPolicyResp {
	return challengeResponseMapperInst.ToTopologyTrafficPolicyResps(policies)
}

func ChallengeImportTopologyRespFromParsed(item *ParsedChallengePackageTopology) *challengecontracts.ChallengeImportTopologyResp {
	if item == nil {
		return nil
	}
	nodes := challengeResponseMapperInst.ToChallengeImportTopologyNodeRespBases(item.Nodes)
	for idx := range nodes {
		nodes[idx].ImageRef = strings.TrimSpace(nodes[idx].ImageRef)
	}
	return &challengecontracts.ChallengeImportTopologyResp{
		Source:       item.Source,
		EntryNodeKey: item.EntryNodeKey,
		Networks:     challengeResponseMapperInst.ToTopologyNetworkResps(importedTopologyNetworkList(item.Networks)),
		Nodes:        nodes,
		Links:        challengeResponseMapperInst.ToTopologyLinkResps(importedTopologyLinkList(item.Links)),
		Policies:     challengeResponseMapperInst.ToTopologyTrafficPolicyResps(importedTopologyPolicyList(item.Policies)),
	}
}

func ChallengePackageFileRespList(items []ParsedChallengePackageFile) []challengecontracts.ChallengePackageFileResp {
	return challengeResponseMapperInst.ToChallengePackageFileResps(items)
}

func ChallengePackageFileRespListFromRevisionFiles(items []challengecontracts.ChallengePackageFileResp) []challengecontracts.ChallengePackageFileResp {
	return append([]challengecontracts.ChallengePackageFileResp(nil), items...)
}

func topologySpecRespFromSpec(entryNodeKey string, spec challengecontracts.TopologySpec) *challengecontracts.TopologySpecResp {
	return &challengecontracts.TopologySpecResp{
		EntryNodeKey: entryNodeKey,
		Networks:     topologyNetworkRespList(spec.Networks),
		Nodes:        topologyNodeRespList(spec.Nodes),
		Links:        topologyLinkRespList(spec.Links),
		Policies:     topologyTrafficPolicyRespList(spec.Policies),
	}
}

func importedTopologyNetworkList(items []ChallengePackageTopologyNetwork) []challengecontracts.TopologyNetwork {
	return challengeResponseMapperInst.ToImportedTopologyNetworks(items)
}

func importedTopologyLinkList(items []ChallengePackageTopologyLink) []challengecontracts.TopologyLink {
	return challengeResponseMapperInst.ToImportedTopologyLinks(items)
}

func importedTopologyPolicyList(items []ChallengePackageTopologyPolicy) []challengecontracts.TopologyTrafficPolicy {
	return challengeResponseMapperInst.ToImportedTopologyPolicies(items)
}

func normalizeTopologyNetworks(networks []challengecontracts.TopologyNetworkReq) ([]challengecontracts.TopologyNetwork, map[string]struct{}, string, error) {
	if len(networks) == 0 {
		return []challengecontracts.TopologyNetwork{
				{
					Key:  challengecontracts.TopologyDefaultNetworkKey,
					Name: challengecontracts.TopologyDefaultNetworkKey,
				},
			},
			map[string]struct{}{challengecontracts.TopologyDefaultNetworkKey: {}},
			challengecontracts.TopologyDefaultNetworkKey,
			nil
	}

	seen := make(map[string]struct{}, len(networks))
	items := make([]challengecontracts.TopologyNetwork, 0, len(networks))
	fallbackNetworkKey := ""
	for _, network := range networks {
		key := strings.TrimSpace(network.Key)
		if key == "" {
			return nil, nil, "", apperror.ErrInvalidParams.WithCause(errors.New("网络 key 不能为空"))
		}
		if _, exists := seen[key]; exists {
			return nil, nil, "", apperror.ErrInvalidParams.WithCause(errors.New("网络 key 不能重复"))
		}
		if fallbackNetworkKey == "" {
			fallbackNetworkKey = key
		}
		seen[key] = struct{}{}
		items = append(items, challengecontracts.TopologyNetwork{
			Key:      key,
			Name:     strings.TrimSpace(network.Name),
			CIDR:     strings.TrimSpace(network.CIDR),
			Internal: network.Internal,
		})
	}
	return items, seen, fallbackNetworkKey, nil
}

func normalizeNodeNetworks(networkKeys []string, validNetworkKeys map[string]struct{}, fallbackNetworkKey string) ([]string, error) {
	if len(validNetworkKeys) == 1 && len(networkKeys) == 0 {
		return []string{fallbackNetworkKey}, nil
	}
	if len(networkKeys) == 0 {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("多网络拓扑中的节点必须声明 network_keys"))
	}

	seen := make(map[string]struct{}, len(networkKeys))
	items := make([]string, 0, len(networkKeys))
	for _, networkKey := range networkKeys {
		key := strings.TrimSpace(networkKey)
		if key == "" {
			return nil, apperror.ErrInvalidParams.WithCause(errors.New("节点 network_keys 不能为空"))
		}
		if _, exists := validNetworkKeys[key]; !exists {
			return nil, apperror.ErrInvalidParams.WithCause(errors.New("节点引用了不存在的网络"))
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, key)
	}
	return items, nil
}

func normalizePolicyProtocol(protocol string) string {
	trimmed := strings.TrimSpace(protocol)
	if trimmed == "" {
		return challengecontracts.TopologyPolicyProtocolAny
	}
	return trimmed
}

func normalizeServiceProtocol(protocol string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case runtimecontracts.ChallengeTargetProtocolTCP:
		return runtimecontracts.ChallengeTargetProtocolTCP
	default:
		return runtimecontracts.ChallengeTargetProtocolHTTP
	}
}

func normalizePolicyPorts(ports []int) []int {
	if len(ports) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ports))
	items := make([]int, 0, len(ports))
	for _, port := range ports {
		if _, exists := seen[port]; exists {
			continue
		}
		seen[port] = struct{}{}
		items = append(items, port)
	}
	return items
}

func trimEnvMap(env map[string]string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	result := make(map[string]string, len(env))
	for key, value := range env {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		result[trimmedKey] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
