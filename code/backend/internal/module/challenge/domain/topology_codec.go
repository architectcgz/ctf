package domain

import (
	"errors"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func BuildTopologySpec(entryNodeKey string, networks []dto.TopologyNetworkReq, nodes []dto.TopologyNodeReq, links []dto.TopologyLinkReq, policies []dto.TopologyTrafficPolicyReq) (string, string, error) {
	spec, normalizedEntryNodeKey, err := normalizeTopologySpec(entryNodeKey, networks, nodes, links, policies)
	if err != nil {
		return "", "", err
	}
	raw, err := model.EncodeTopologySpec(spec)
	if err != nil {
		return "", "", err
	}
	return raw, normalizedEntryNodeKey, nil
}

func normalizeTopologySpec(entryNodeKey string, networks []dto.TopologyNetworkReq, nodes []dto.TopologyNodeReq, links []dto.TopologyLinkReq, policies []dto.TopologyTrafficPolicyReq) (model.TopologySpec, string, error) {
	if len(nodes) == 0 {
		return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("拓扑至少需要一个节点"))
	}

	specNetworks, networkKeys, fallbackNetworkKey, err := normalizeTopologyNetworks(networks)
	if err != nil {
		return model.TopologySpec{}, "", err
	}

	seenNodes := make(map[string]struct{}, len(nodes))
	specNodes := make([]model.TopologyNode, 0, len(nodes))
	injectFlagCount := 0
	for _, node := range nodes {
		key := strings.TrimSpace(node.Key)
		if key == "" {
			return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("节点 key 不能为空"))
		}
		if _, exists := seenNodes[key]; exists {
			return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("节点 key 不能重复"))
		}
		seenNodes[key] = struct{}{}

		tier := strings.TrimSpace(node.Tier)
		if tier == "" {
			tier = model.TopologyTierService
		}
		networkRefs, networkErr := normalizeNodeNetworks(node.NetworkKeys, networkKeys, fallbackNetworkKey)
		if networkErr != nil {
			return model.TopologySpec{}, "", networkErr
		}

		var resources *model.TopologyResources
		if node.Resources != nil {
			resources = &model.TopologyResources{
				CPUQuota:  node.Resources.CPUQuota,
				MemoryMB:  node.Resources.MemoryMB,
				PidsLimit: node.Resources.PidsLimit,
			}
		}
		if node.InjectFlag {
			injectFlagCount++
		}

		specNodes = append(specNodes, model.TopologyNode{
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
		return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("入口节点不存在"))
	}

	if injectFlagCount == 0 {
		for idx := range specNodes {
			if specNodes[idx].Key == entry {
				specNodes[idx].InjectFlag = true
				break
			}
		}
	}

	specLinks := make([]model.TopologyLink, 0, len(links))
	for _, link := range links {
		from := strings.TrimSpace(link.FromNodeKey)
		to := strings.TrimSpace(link.ToNodeKey)
		if _, exists := seenNodes[from]; !exists {
			return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("拓扑连线源节点不存在"))
		}
		if _, exists := seenNodes[to]; !exists {
			return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("拓扑连线目标节点不存在"))
		}
		specLinks = append(specLinks, model.TopologyLink{
			FromNodeKey: from,
			ToNodeKey:   to,
		})
	}

	specPolicies := make([]model.TopologyTrafficPolicy, 0, len(policies))
	for _, policy := range policies {
		source := strings.TrimSpace(policy.SourceNodeKey)
		target := strings.TrimSpace(policy.TargetNodeKey)
		if _, exists := seenNodes[source]; !exists {
			return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("链路策略源节点不存在"))
		}
		if _, exists := seenNodes[target]; !exists {
			return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("链路策略目标节点不存在"))
		}
		specPolicies = append(specPolicies, model.TopologyTrafficPolicy{
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
		return model.TopologySpec{}, "", errcode.ErrInvalidParams.WithCause(errors.New("入口节点必须配置 service_port"))
	}

	return model.TopologySpec{
		Networks: specNetworks,
		Nodes:    specNodes,
		Links:    specLinks,
		Policies: specPolicies,
	}, entry, nil
}

func TopologyRespFromModel(item *model.ChallengeTopology) (*dto.ChallengeTopologyResp, error) {
	spec, err := model.DecodeTopologySpec(item.Spec)
	if err != nil {
		return nil, err
	}
	var baseline *dto.TopologySpecResp
	if strings.TrimSpace(item.PackageBaselineSpec) != "" {
		baselineSpec, decodeErr := model.DecodeTopologySpec(item.PackageBaselineSpec)
		if decodeErr != nil {
			return nil, decodeErr
		}
		baseline = topologySpecRespFromSpec(item.EntryNodeKey, baselineSpec)
	}
	return &dto.ChallengeTopologyResp{
		ID:                   item.ID,
		ChallengeID:          item.ChallengeID,
		TemplateID:           item.TemplateID,
		EntryNodeKey:         item.EntryNodeKey,
		Networks:             topologyNetworkRespList(spec.Networks),
		Nodes:                topologyNodeRespList(spec.Nodes),
		Links:                topologyLinkRespList(spec.Links),
		Policies:             topologyTrafficPolicyRespList(spec.Policies),
		SourceType:           item.SourceType,
		SourcePath:           item.SourcePath,
		SyncStatus:           item.SyncStatus,
		PackageRevisionID:    item.PackageRevisionID,
		LastExportRevisionID: item.LastExportRevisionID,
		PackageBaseline:      baseline,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}, nil
}

func TopologySpecRespFromEncoded(entryNodeKey string, raw string) (*dto.TopologySpecResp, error) {
	spec, err := model.DecodeTopologySpec(raw)
	if err != nil {
		return nil, err
	}
	return topologySpecRespFromSpec(entryNodeKey, spec), nil
}

func TopologySpecRespFromSpec(entryNodeKey string, spec model.TopologySpec) *dto.TopologySpecResp {
	return topologySpecRespFromSpec(entryNodeKey, spec)
}

func TemplateRespFromModel(item *model.EnvironmentTemplate) (*dto.EnvironmentTemplateResp, error) {
	spec, err := model.DecodeTopologySpec(item.Spec)
	if err != nil {
		return nil, err
	}
	return &dto.EnvironmentTemplateResp{
		ID:           item.ID,
		Name:         item.Name,
		Description:  item.Description,
		EntryNodeKey: item.EntryNodeKey,
		Networks:     topologyNetworkRespList(spec.Networks),
		Nodes:        topologyNodeRespList(spec.Nodes),
		Links:        topologyLinkRespList(spec.Links),
		Policies:     topologyTrafficPolicyRespList(spec.Policies),
		UsageCount:   item.UsageCount,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}, nil
}

func topologyNodeRespList(nodes []model.TopologyNode) []dto.TopologyNodeResp {
	items := make([]dto.TopologyNodeResp, 0, len(nodes))
	for _, node := range nodes {
		var resources *dto.TopologyResourcesResp
		if node.Resources != nil {
			resources = &dto.TopologyResourcesResp{
				CPUQuota:  node.Resources.CPUQuota,
				MemoryMB:  node.Resources.MemoryMB,
				PidsLimit: node.Resources.PidsLimit,
			}
		}
		items = append(items, dto.TopologyNodeResp{
			Key:             node.Key,
			Name:            node.Name,
			ImageID:         node.ImageID,
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			InjectFlag:      node.InjectFlag,
			Tier:            node.Tier,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			Env:             node.Env,
			Resources:       resources,
		})
	}
	return items
}

func topologyNetworkRespList(networks []model.TopologyNetwork) []dto.TopologyNetworkResp {
	items := make([]dto.TopologyNetworkResp, 0, len(networks))
	for _, network := range networks {
		items = append(items, dto.TopologyNetworkResp{
			Key:      network.Key,
			Name:     network.Name,
			CIDR:     network.CIDR,
			Internal: network.Internal,
		})
	}
	return items
}

func topologyLinkRespList(links []model.TopologyLink) []dto.TopologyLinkResp {
	items := make([]dto.TopologyLinkResp, 0, len(links))
	for _, link := range links {
		items = append(items, dto.TopologyLinkResp{
			FromNodeKey: link.FromNodeKey,
			ToNodeKey:   link.ToNodeKey,
		})
	}
	return items
}

func topologyTrafficPolicyRespList(policies []model.TopologyTrafficPolicy) []dto.TopologyTrafficPolicyResp {
	items := make([]dto.TopologyTrafficPolicyResp, 0, len(policies))
	for _, policy := range policies {
		items = append(items, dto.TopologyTrafficPolicyResp{
			SourceNodeKey: policy.SourceNodeKey,
			TargetNodeKey: policy.TargetNodeKey,
			Action:        policy.Action,
			Protocol:      policy.Protocol,
			Ports:         append([]int(nil), policy.Ports...),
		})
	}
	return items
}

func ChallengePackageRevisionRespFromModel(item *model.ChallengePackageRevision) dto.ChallengePackageRevisionResp {
	return dto.ChallengePackageRevisionResp{
		ID:                 item.ID,
		RevisionNo:         item.RevisionNo,
		SourceType:         item.SourceType,
		ParentRevisionID:   item.ParentRevisionID,
		PackageSlug:        item.PackageSlug,
		ArchivePath:        item.ArchivePath,
		SourceDir:          item.SourceDir,
		TopologySourcePath: item.TopologySourcePath,
		CreatedBy:          item.CreatedBy,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func ChallengeImportTopologyRespFromParsed(item *ParsedChallengePackageTopology) *dto.ChallengeImportTopologyResp {
	if item == nil {
		return nil
	}
	nodes := make([]dto.ChallengeImportTopologyNodeResp, 0, len(item.Nodes))
	for _, node := range item.Nodes {
		nodes = append(nodes, dto.ChallengeImportTopologyNodeResp{
			Key:         node.Key,
			Name:        node.Name,
			ImageRef:    strings.TrimSpace(node.Image.Ref),
			ServicePort: node.ServicePort,
			InjectFlag:  node.InjectFlag,
			Tier:        node.Tier,
			NetworkKeys: append([]string(nil), node.NetworkKeys...),
			Env:         node.Env,
		})
	}
	return &dto.ChallengeImportTopologyResp{
		Source:       item.Source,
		EntryNodeKey: item.EntryNodeKey,
		Networks:     topologyNetworkRespList(importedTopologyNetworkList(item.Networks)),
		Nodes:        nodes,
		Links:        topologyLinkRespList(importedTopologyLinkList(item.Links)),
		Policies:     topologyTrafficPolicyRespList(importedTopologyPolicyList(item.Policies)),
	}
}

func ChallengePackageFileRespList(items []ParsedChallengePackageFile) []dto.ChallengePackageFileResp {
	resp := make([]dto.ChallengePackageFileResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.ChallengePackageFileResp{
			Path: item.Path,
			Size: item.Size,
		})
	}
	return resp
}

func ChallengePackageFileRespListFromRevisionFiles(items []dto.ChallengePackageFileResp) []dto.ChallengePackageFileResp {
	return append([]dto.ChallengePackageFileResp(nil), items...)
}

func topologySpecRespFromSpec(entryNodeKey string, spec model.TopologySpec) *dto.TopologySpecResp {
	return &dto.TopologySpecResp{
		EntryNodeKey: entryNodeKey,
		Networks:     topologyNetworkRespList(spec.Networks),
		Nodes:        topologyNodeRespList(spec.Nodes),
		Links:        topologyLinkRespList(spec.Links),
		Policies:     topologyTrafficPolicyRespList(spec.Policies),
	}
}

func importedTopologyNetworkList(items []ChallengePackageTopologyNetwork) []model.TopologyNetwork {
	resp := make([]model.TopologyNetwork, 0, len(items))
	for _, item := range items {
		resp = append(resp, model.TopologyNetwork{
			Key:      item.Key,
			Name:     item.Name,
			CIDR:     item.CIDR,
			Internal: item.Internal,
		})
	}
	return resp
}

func importedTopologyLinkList(items []ChallengePackageTopologyLink) []model.TopologyLink {
	resp := make([]model.TopologyLink, 0, len(items))
	for _, item := range items {
		resp = append(resp, model.TopologyLink{
			FromNodeKey: item.FromNodeKey,
			ToNodeKey:   item.ToNodeKey,
		})
	}
	return resp
}

func importedTopologyPolicyList(items []ChallengePackageTopologyPolicy) []model.TopologyTrafficPolicy {
	resp := make([]model.TopologyTrafficPolicy, 0, len(items))
	for _, item := range items {
		resp = append(resp, model.TopologyTrafficPolicy{
			SourceNodeKey: item.SourceNodeKey,
			TargetNodeKey: item.TargetNodeKey,
			Action:        item.Action,
			Protocol:      item.Protocol,
			Ports:         append([]int(nil), item.Ports...),
		})
	}
	return resp
}

func normalizeTopologyNetworks(networks []dto.TopologyNetworkReq) ([]model.TopologyNetwork, map[string]struct{}, string, error) {
	if len(networks) == 0 {
		return []model.TopologyNetwork{
				{
					Key:  model.TopologyDefaultNetworkKey,
					Name: model.TopologyDefaultNetworkKey,
				},
			},
			map[string]struct{}{model.TopologyDefaultNetworkKey: {}},
			model.TopologyDefaultNetworkKey,
			nil
	}

	seen := make(map[string]struct{}, len(networks))
	items := make([]model.TopologyNetwork, 0, len(networks))
	fallbackNetworkKey := ""
	for _, network := range networks {
		key := strings.TrimSpace(network.Key)
		if key == "" {
			return nil, nil, "", errcode.ErrInvalidParams.WithCause(errors.New("网络 key 不能为空"))
		}
		if _, exists := seen[key]; exists {
			return nil, nil, "", errcode.ErrInvalidParams.WithCause(errors.New("网络 key 不能重复"))
		}
		if fallbackNetworkKey == "" {
			fallbackNetworkKey = key
		}
		seen[key] = struct{}{}
		items = append(items, model.TopologyNetwork{
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
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("多网络拓扑中的节点必须声明 network_keys"))
	}

	seen := make(map[string]struct{}, len(networkKeys))
	items := make([]string, 0, len(networkKeys))
	for _, networkKey := range networkKeys {
		key := strings.TrimSpace(networkKey)
		if key == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("节点 network_keys 不能为空"))
		}
		if _, exists := validNetworkKeys[key]; !exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("节点引用了不存在的网络"))
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
		return model.TopologyPolicyProtocolAny
	}
	return trimmed
}

func normalizeServiceProtocol(protocol string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case model.ChallengeTargetProtocolTCP:
		return model.ChallengeTargetProtocolTCP
	default:
		return model.ChallengeTargetProtocolHTTP
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
