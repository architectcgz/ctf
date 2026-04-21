package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"gopkg.in/yaml.v3"
)

func parseChallengePackageTopology(
	rootDir string,
	extension ChallengePackageTopologyExtension,
) (*ParsedChallengePackageTopology, error) {
	if !extension.Enabled {
		return nil, nil
	}

	source := strings.TrimSpace(extension.Source)
	if source == "" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("extensions.topology.source 不能为空"))
	}

	topologyPath, err := safePackageJoin(rootDir, source)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("拓扑文件路径非法: %w", err))
	}
	content, err := os.ReadFile(topologyPath)
	if err != nil {
		return nil, fmt.Errorf("read topology %s: %w", topologyPath, err)
	}

	var manifest ChallengePackageTopologyManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parse topology %s: %w", topologyPath, err)
	}
	if strings.TrimSpace(manifest.APIVersion) != "v1" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("topology.yml api_version 仅支持 v1"))
	}
	if strings.TrimSpace(manifest.Kind) != "topology" {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("topology.yml kind 必须为 topology"))
	}
	if len(manifest.Nodes) == 0 {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑至少需要一个节点"))
	}

	parsed := &ParsedChallengePackageTopology{
		Source:       filepath.ToSlash(filepath.Clean(source)),
		Raw:          string(content),
		EntryNodeKey: strings.TrimSpace(manifest.EntryNodeKey),
		Networks:     make([]ChallengePackageTopologyNetwork, 0, len(manifest.Networks)),
		Nodes:        make([]ChallengePackageTopologyNode, 0, len(manifest.Nodes)),
		Links:        make([]ChallengePackageTopologyLink, 0, len(manifest.Links)),
		Policies:     make([]ChallengePackageTopologyPolicy, 0, len(manifest.Policies)),
	}

	networkKeys := map[string]struct{}{}
	for _, network := range manifest.Networks {
		key := strings.TrimSpace(network.Key)
		if key == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("拓扑网络 key 不能为空"))
		}
		if _, exists := networkKeys[key]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("拓扑网络 key 重复: %s", key))
		}
		networkKeys[key] = struct{}{}
		parsed.Networks = append(parsed.Networks, ChallengePackageTopologyNetwork{
			Key:      key,
			Name:     strings.TrimSpace(network.Name),
			CIDR:     strings.TrimSpace(network.CIDR),
			Internal: network.Internal,
		})
	}
	if len(parsed.Networks) == 0 {
		parsed.Networks = append(parsed.Networks, ChallengePackageTopologyNetwork{
			Key:  model.TopologyDefaultNetworkKey,
			Name: model.TopologyDefaultNetworkKey,
		})
		networkKeys[model.TopologyDefaultNetworkKey] = struct{}{}
	}

	nodeKeys := map[string]struct{}{}
	injectFlagCount := 0
	for _, node := range manifest.Nodes {
		key := strings.TrimSpace(node.Key)
		if key == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("拓扑节点 key 不能为空"))
		}
		if _, exists := nodeKeys[key]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("拓扑节点 key 重复: %s", key))
		}
		nodeKeys[key] = struct{}{}
		if strings.TrimSpace(node.Image.Ref) == "" {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("拓扑节点 %s 缺少 image.ref", key))
		}
		if node.InjectFlag {
			injectFlagCount++
		}
		networkRefs, err := normalizeImportedTopologyNodeNetworks(node.NetworkKeys, networkKeys)
		if err != nil {
			return nil, err
		}
		parsed.Nodes = append(parsed.Nodes, ChallengePackageTopologyNode{
			Key:         key,
			Name:        strings.TrimSpace(node.Name),
			Tier:        strings.TrimSpace(node.Tier),
			Image:       node.Image,
			ServicePort: node.ServicePort,
			InjectFlag:  node.InjectFlag,
			NetworkKeys: networkRefs,
			Env:         trimEnvMap(node.Env),
			Resources:   node.Resources,
		})
	}

	if parsed.EntryNodeKey == "" {
		parsed.EntryNodeKey = parsed.Nodes[0].Key
	}
	if _, exists := nodeKeys[parsed.EntryNodeKey]; !exists {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑入口节点不存在"))
	}

	entryHasPort := false
	for idx := range parsed.Nodes {
		if parsed.Nodes[idx].Key != parsed.EntryNodeKey {
			continue
		}
		if injectFlagCount == 0 {
			parsed.Nodes[idx].InjectFlag = true
		}
		if parsed.Nodes[idx].ServicePort > 0 {
			entryHasPort = true
		}
	}
	if !entryHasPort {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑入口节点必须配置 service_port"))
	}

	for _, link := range manifest.Links {
		from := strings.TrimSpace(link.FromNodeKey)
		to := strings.TrimSpace(link.ToNodeKey)
		if _, exists := nodeKeys[from]; !exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑连线源节点不存在"))
		}
		if _, exists := nodeKeys[to]; !exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑连线目标节点不存在"))
		}
		parsed.Links = append(parsed.Links, ChallengePackageTopologyLink{
			FromNodeKey: from,
			ToNodeKey:   to,
		})
	}

	for _, policy := range manifest.Policies {
		sourceKey := strings.TrimSpace(policy.SourceNodeKey)
		targetKey := strings.TrimSpace(policy.TargetNodeKey)
		if _, exists := nodeKeys[sourceKey]; !exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑策略源节点不存在"))
		}
		if _, exists := nodeKeys[targetKey]; !exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑策略目标节点不存在"))
		}
		action := strings.TrimSpace(policy.Action)
		switch action {
		case model.TopologyPolicyActionAllow, model.TopologyPolicyActionDeny:
		default:
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑策略 action 仅支持 allow/deny"))
		}
		parsed.Policies = append(parsed.Policies, ChallengePackageTopologyPolicy{
			SourceNodeKey: sourceKey,
			TargetNodeKey: targetKey,
			Action:        action,
			Protocol:      normalizePolicyProtocol(policy.Protocol),
			Ports:         normalizePolicyPorts(policy.Ports),
		})
	}

	return parsed, nil
}

func normalizeImportedTopologyNodeNetworks(
	networkKeys []string,
	knownNetworks map[string]struct{},
) ([]string, error) {
	if len(networkKeys) == 0 {
		return []string{model.TopologyDefaultNetworkKey}, nil
	}
	result := make([]string, 0, len(networkKeys))
	seen := map[string]struct{}{}
	for _, item := range networkKeys {
		key := strings.TrimSpace(item)
		if key == "" {
			continue
		}
		if _, exists := knownNetworks[key]; !exists {
			return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("节点引用未知网络: %s", key))
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	if len(result) == 0 {
		return []string{model.TopologyDefaultNetworkKey}, nil
	}
	sort.Strings(result)
	return result, nil
}

func listChallengePackageFiles(rootDir string) ([]ParsedChallengePackageFile, error) {
	files := make([]ParsedChallengePackageFile, 0, 16)
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		files = append(files, ParsedChallengePackageFile{
			Path: filepath.ToSlash(filepath.Clean(relative)),
			Size: info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func BuildTopologySpecFromImportedPackage(
	topology *ParsedChallengePackageTopology,
	resolveImageID func(imageRef string) (int64, error),
) (string, string, error) {
	if topology == nil {
		return "", "", errcode.ErrInvalidParams.WithCause(errors.New("题目包拓扑不能为空"))
	}
	if resolveImageID == nil {
		return "", "", errcode.ErrInvalidParams.WithCause(errors.New("缺少拓扑镜像解析器"))
	}

	specNetworks, networkKeys, fallbackNetworkKey, err := normalizeTopologyNetworks(toTopologyNetworkReqs(topology.Networks))
	if err != nil {
		return "", "", err
	}

	seenNodes := make(map[string]struct{}, len(topology.Nodes))
	specNodes := make([]model.TopologyNode, 0, len(topology.Nodes))
	for _, node := range topology.Nodes {
		key := strings.TrimSpace(node.Key)
		if _, exists := seenNodes[key]; exists {
			return "", "", errcode.ErrInvalidParams.WithCause(errors.New("节点 key 不能重复"))
		}
		seenNodes[key] = struct{}{}

		imageID, err := resolveImageID(strings.TrimSpace(node.Image.Ref))
		if err != nil {
			return "", "", err
		}
		networkRefs, err := normalizeNodeNetworks(node.NetworkKeys, networkKeys, fallbackNetworkKey)
		if err != nil {
			return "", "", err
		}
		tier := strings.TrimSpace(node.Tier)
		if tier == "" {
			tier = model.TopologyTierService
		}
		specNode := model.TopologyNode{
			Key:         key,
			Name:        strings.TrimSpace(node.Name),
			ImageID:     imageID,
			ServicePort: node.ServicePort,
			InjectFlag:  node.InjectFlag,
			Tier:        tier,
			NetworkKeys: networkRefs,
			Env:         trimEnvMap(node.Env),
		}
		if node.Resources != nil {
			specNode.Resources = &model.TopologyResources{
				CPUQuota:  node.Resources.CPUQuota,
				MemoryMB:  node.Resources.MemoryMB,
				PidsLimit: node.Resources.PidsLimit,
			}
		}
		specNodes = append(specNodes, specNode)
	}

	specLinks := make([]model.TopologyLink, 0, len(topology.Links))
	for _, link := range topology.Links {
		specLinks = append(specLinks, model.TopologyLink{
			FromNodeKey: strings.TrimSpace(link.FromNodeKey),
			ToNodeKey:   strings.TrimSpace(link.ToNodeKey),
		})
	}

	specPolicies := make([]model.TopologyTrafficPolicy, 0, len(topology.Policies))
	for _, policy := range topology.Policies {
		specPolicies = append(specPolicies, model.TopologyTrafficPolicy{
			SourceNodeKey: strings.TrimSpace(policy.SourceNodeKey),
			TargetNodeKey: strings.TrimSpace(policy.TargetNodeKey),
			Action:        strings.TrimSpace(policy.Action),
			Protocol:      normalizePolicyProtocol(policy.Protocol),
			Ports:         normalizePolicyPorts(policy.Ports),
		})
	}

	raw, err := model.EncodeTopologySpec(model.TopologySpec{
		Networks: specNetworks,
		Nodes:    specNodes,
		Links:    specLinks,
		Policies: specPolicies,
	})
	if err != nil {
		return "", "", err
	}
	return raw, strings.TrimSpace(topology.EntryNodeKey), nil
}

func toTopologyNetworkReqs(items []ChallengePackageTopologyNetwork) []dto.TopologyNetworkReq {
	reqs := make([]dto.TopologyNetworkReq, 0, len(items))
	for _, item := range items {
		reqs = append(reqs, dto.TopologyNetworkReq{
			Key:      item.Key,
			Name:     item.Name,
			CIDR:     item.CIDR,
			Internal: item.Internal,
		})
	}
	return reqs
}
