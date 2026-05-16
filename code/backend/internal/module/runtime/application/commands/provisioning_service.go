package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

const (
	managedContainerNamePrefix = "ctf-instance-"
	managedNetworkNamePrefix   = "ctf-net-"
	awdContestNetworkPrefix    = "ctf-awd-contest-"
	awdWorkspaceNamePrefix     = "ctf-workspace-"
)

type provisioningRepository interface {
	ReserveAvailablePort(ctx context.Context, start, end int) (int, error)
	ReleaseReservedPort(ctx context.Context, port int) error
}

type createdTopologyNetwork struct {
	key      string
	name     string
	id       string
	internal bool
	shared   bool
}

// ProvisioningService 收口运行时资源创建编排，包括单容器与拓扑实例创建。
type ProvisioningService struct {
	repo   provisioningRepository
	engine runtimeports.ContainerProvisioningRuntime
	config *config.ContainerConfig
	logger *zap.Logger
}

// NewProvisioningService 创建运行时资源编排服务。
func NewProvisioningService(repo provisioningRepository, engine runtimeports.ContainerProvisioningRuntime, cfg *config.ContainerConfig, logger *zap.Logger) *ProvisioningService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if isNilCommandDependency(repo) {
		repo = nil
	}
	if isNilCommandDependency(engine) {
		engine = nil
	}
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}
	return &ProvisioningService{
		repo:   repo,
		engine: engine,
		config: cfg,
		logger: logger,
	}
}

// CreateContainer 为单容器题目创建默认拓扑，并返回入口容器与端口信息。
func (s *ProvisioningService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	servicePort, err = s.resolveServicePort(ctx, imageName)
	if err != nil {
		return "", "", 0, 0, err
	}

	result, err := s.CreateTopology(ctx, &runtimeports.TopologyCreateRequest{
		ReservedHostPort: reservedHostPort,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:             "default",
				Image:           imageName,
				Env:             env,
				ServicePort:     servicePort,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
			},
		},
	})
	if err != nil {
		return "", "", 0, 0, err
	}

	for _, item := range result.RuntimeDetails.Containers {
		if item.IsEntryPoint {
			hostPort = item.HostPort
			break
		}
	}
	return result.PrimaryContainerID, result.NetworkID, hostPort, servicePort, nil
}

// CreateTopology 按拓扑请求创建网络、容器与 ACL 规则。
func (s *ProvisioningService) CreateTopology(ctx context.Context, req *runtimeports.TopologyCreateRequest) (*runtimeports.TopologyCreateResult, error) {
	ctx = normalizeContext(ctx)
	if req == nil || len(req.Nodes) == 0 {
		return nil, fmt.Errorf("topology nodes are required")
	}
	if s.engine == nil {
		return nil, fmt.Errorf("runtime engine is not configured")
	}

	networks := normalizedCreateNetworks(req.Networks)
	entryNodeIndex := -1
	for idx, node := range req.Nodes {
		if node.IsEntryPoint {
			entryNodeIndex = idx
			break
		}
	}
	if entryNodeIndex < 0 {
		return nil, fmt.Errorf("entry node is required")
	}

	publishEntryPort := !req.DisableEntryPortPublishing
	hostPort := req.ReservedHostPort
	allocatedHostPort := 0
	success := false
	if publishEntryPort && hostPort <= 0 {
		var err error
		hostPort, err = s.allocatePort(ctx)
		if err != nil {
			return nil, err
		}
		allocatedHostPort = hostPort
	}
	defer func() {
		if !success && allocatedHostPort > 0 {
			_ = s.repo.ReleaseReservedPort(ctx, allocatedHostPort)
		}
	}()

	createdNetworks := make([]createdTopologyNetwork, 0, len(networks))
	networkByKey := make(map[string]createdTopologyNetwork, len(networks))
	managedLabels := managedContainerLabels(req)
	for _, network := range networks {
		networkName := resolveCreateNetworkName(network)
		networkID, err := s.engine.CreateNetwork(ctx, networkName, managedLabels, network.Internal, network.Shared)
		if err != nil {
			s.cleanupTopologyResources(ctx, nil, collectOwnedNetworkIDs(createdNetworks))
			return nil, err
		}
		item := createdTopologyNetwork{
			key:      network.Key,
			name:     networkName,
			id:       networkID,
			internal: network.Internal,
			shared:   network.Shared,
		}
		createdNetworks = append(createdNetworks, item)
		networkByKey[network.Key] = item
	}

	details := model.InstanceRuntimeDetails{
		Networks:   make([]model.InstanceRuntimeNetwork, 0, len(createdNetworks)),
		Containers: make([]model.InstanceRuntimeContainer, 0, len(req.Nodes)),
	}
	for _, network := range createdNetworks {
		details.Networks = append(details.Networks, model.InstanceRuntimeNetwork{
			Key:       network.key,
			Name:      network.name,
			NetworkID: network.id,
			Internal:  network.internal,
			Shared:    network.shared,
		})
	}

	createdContainerIDs := make([]string, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodeNetworkKeys := normalizedNodeNetworkKeys(node.NetworkKeys, networks)
		primaryNetwork := networkByKey[nodeNetworkKeys[0]]
		servicePort := node.ServicePort
		if node.IsEntryPoint && servicePort <= 0 {
			resolvedPort, err := s.resolveServicePort(ctx, node.Image)
			if err != nil {
				s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
				return nil, err
			}
			servicePort = resolvedPort
		}
		ports := map[string]string(nil)
		if node.IsEntryPoint && publishEntryPort {
			ports = map[string]string{
				strconv.Itoa(servicePort): strconv.Itoa(hostPort),
			}
		}

		containerID, err := s.engine.CreateContainer(ctx, &model.ContainerConfig{
			Image:          node.Image,
			Name:           buildManagedContainerName(req.ContainerName),
			Env:            envMapToList(node.Env),
			Command:        append([]string(nil), node.Command...),
			WorkingDir:     strings.TrimSpace(node.WorkingDir),
			Ports:          ports,
			Mounts:         append([]model.ContainerMount(nil), node.Mounts...),
			Labels:         managedLabels,
			Resources:      node.Resources,
			Network:        primaryNetwork.name,
			NetworkAliases: normalizedNetworkAliases(node.NetworkAliases),
		})
		if err != nil {
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
			return nil, err
		}
		if err := s.engine.StartContainer(ctx, containerID); err != nil {
			createdContainerIDs = append(createdContainerIDs, containerID)
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
			return nil, err
		}
		for _, networkKey := range nodeNetworkKeys[1:] {
			if err := s.engine.ConnectContainerToNetwork(ctx, containerID, networkByKey[networkKey].name); err != nil {
				createdContainerIDs = append(createdContainerIDs, containerID)
				s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
				return nil, err
			}
		}
		networkIPs, err := s.engine.InspectContainerNetworkIPs(ctx, containerID)
		if err != nil {
			createdContainerIDs = append(createdContainerIDs, containerID)
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
			return nil, err
		}

		createdContainerIDs = append(createdContainerIDs, containerID)
		serviceProtocol := normalizeServiceProtocol(node.ServiceProtocol)
		runtimeItem := model.InstanceRuntimeContainer{
			NodeKey:         node.Key,
			ContainerID:     containerID,
			ServicePort:     servicePort,
			ServiceProtocol: serviceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), nodeNetworkKeys...),
			NetworkAliases:  normalizedNetworkAliases(node.NetworkAliases),
			NetworkIPs:      networkIPs,
		}
		if node.IsEntryPoint && publishEntryPort {
			runtimeItem.HostPort = hostPort
		}
		details.Containers = append(details.Containers, runtimeItem)
	}

	resolvedACLRules, err := s.resolveTopologyACLRules(ctx, req, details)
	if err != nil {
		s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
		return nil, err
	}
	if len(resolvedACLRules) > 0 {
		if err := s.engine.ApplyACLRules(ctx, resolvedACLRules); err != nil {
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
			return nil, err
		}
		details.ACLRules = resolvedACLRules
	}

	accessURL, err := s.resolveEntryAccessURL(ctx, details, entryNodeIndex, publishEntryPort, hostPort)
	if err != nil {
		s.cleanupTopologyResources(ctx, createdContainerIDs, collectOwnedNetworkIDs(createdNetworks))
		return nil, err
	}

	success = true
	return &runtimeports.TopologyCreateResult{
		PrimaryContainerID: details.Containers[entryNodeIndex].ContainerID,
		NetworkID:          details.Networks[0].NetworkID,
		AccessURL:          accessURL,
		RuntimeDetails:     details,
	}, nil
}

func (s *ProvisioningService) resolveEntryAccessURL(ctx context.Context, details model.InstanceRuntimeDetails, entryNodeIndex int, publishEntryPort bool, hostPort int) (string, error) {
	if entryNodeIndex < 0 || entryNodeIndex >= len(details.Containers) {
		return "", fmt.Errorf("entry container is missing")
	}
	entry := details.Containers[entryNodeIndex]
	scheme := normalizeServiceProtocol(entry.ServiceProtocol)
	if publishEntryPort {
		host := model.ResolveRuntimePublishedAccessHost(s.config.PublicHost, s.config.AccessHost)
		return fmt.Sprintf("%s://%s:%d", scheme, host, hostPort), nil
	}
	if entry.ServicePort <= 0 {
		return "", fmt.Errorf("entry service port is required for private access")
	}
	if len(entry.NetworkAliases) > 0 {
		alias := strings.TrimSpace(entry.NetworkAliases[0])
		if alias != "" {
			return fmt.Sprintf("%s://%s:%d", scheme, alias, entry.ServicePort), nil
		}
	}

	ipsByNetworkName, err := s.engine.InspectContainerNetworkIPs(ctx, entry.ContainerID)
	if err != nil {
		return "", err
	}
	networkNamesByKey := make(map[string]string, len(details.Networks))
	for _, network := range details.Networks {
		networkNamesByKey[network.Key] = network.Name
	}
	for _, networkKey := range entry.NetworkKeys {
		networkName := networkNamesByKey[networkKey]
		if networkName == "" {
			continue
		}
		if ip := strings.TrimSpace(ipsByNetworkName[networkName]); ip != "" {
			return fmt.Sprintf("%s://%s:%d", scheme, ip, entry.ServicePort), nil
		}
	}
	for _, ip := range ipsByNetworkName {
		if strings.TrimSpace(ip) != "" {
			return fmt.Sprintf("%s://%s:%d", scheme, strings.TrimSpace(ip), entry.ServicePort), nil
		}
	}
	return "", fmt.Errorf("entry container network ip is not available")
}

func normalizeServiceProtocol(protocol string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case model.ChallengeTargetProtocolTCP:
		return model.ChallengeTargetProtocolTCP
	default:
		return model.ChallengeTargetProtocolHTTP
	}
}

func (s *ProvisioningService) resolveServicePort(ctx context.Context, imageRef string) (int, error) {
	preferredPort := s.config.DefaultExposedPort
	if preferredPort <= 0 {
		preferredPort = 8080
	}
	if s.engine == nil {
		return preferredPort, nil
	}

	resolvedPort, err := s.engine.ResolveServicePort(normalizeContext(ctx), imageRef, preferredPort)
	if err != nil {
		return 0, err
	}
	if resolvedPort <= 0 {
		return preferredPort, nil
	}
	return resolvedPort, nil
}

func (s *ProvisioningService) resolveTopologyACLRules(ctx context.Context, req *runtimeports.TopologyCreateRequest, details model.InstanceRuntimeDetails) ([]model.InstanceRuntimeACLRule, error) {
	if s.engine == nil || req == nil || len(req.Policies) == 0 {
		return nil, nil
	}

	ipsByContainerID := make(map[string]map[string]string, len(details.Containers))
	for _, container := range details.Containers {
		ipsByNetworkName, err := s.engine.InspectContainerNetworkIPs(ctx, container.ContainerID)
		if err != nil {
			return nil, err
		}
		ipsByContainerID[container.ContainerID] = ipsByNetworkName
	}

	return runtimedomain.ResolveTopologyACLRules(req.Policies, details, ipsByContainerID)
}

func (s *ProvisioningService) allocatePort(ctx context.Context) (int, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("runtime provisioning repository is not configured")
	}

	return s.repo.ReserveAvailablePort(ctx, s.config.PortRangeStart, s.config.PortRangeEnd)
}

func (s *ProvisioningService) cleanupTopologyResources(ctx context.Context, containerIDs []string, networkIDs []string) {
	for idx := len(containerIDs) - 1; idx >= 0; idx-- {
		_ = s.removeContainer(ctx, containerIDs[idx])
	}
	for idx := len(networkIDs) - 1; idx >= 0; idx-- {
		_ = s.removeNetwork(ctx, networkIDs[idx])
	}
}

func (s *ProvisioningService) removeContainer(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}
	if s.engine == nil {
		s.logger.Info("删除容器（降级模拟）", zap.String("container_id", containerID))
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(normalizeContext(ctx), 10*time.Second)
	defer cancel()
	_ = s.engine.StopContainer(timeoutCtx, containerID, 5*time.Second)
	if err := s.engine.RemoveContainer(timeoutCtx, containerID, true); err != nil {
		return err
	}

	s.logger.Info("删除容器", zap.String("container_id", containerID))
	return nil
}

func (s *ProvisioningService) removeNetwork(ctx context.Context, networkID string) error {
	if networkID == "" {
		return nil
	}
	if s.engine == nil {
		s.logger.Info("删除网络（降级跳过）", zap.String("network_id", networkID))
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(normalizeContext(ctx), 10*time.Second)
	defer cancel()
	if err := s.engine.RemoveNetwork(timeoutCtx, networkID); err != nil {
		return err
	}

	s.logger.Info("删除网络", zap.String("network_id", networkID))
	return nil
}

func envMapToList(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}

	values := make([]string, 0, len(env))
	for key, value := range env {
		values = append(values, fmt.Sprintf("%s=%s", key, value))
	}
	return values
}

func buildManagedContainerName(preferred string) string {
	preferred = strings.TrimSpace(preferred)
	if preferred != "" {
		return preferred
	}
	return fmt.Sprintf("%s%d", managedContainerNamePrefix, time.Now().UnixNano())
}

func buildManagedNetworkName(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		trimmed = model.TopologyDefaultNetworkKey
	}
	return fmt.Sprintf("%s%s-%d", managedNetworkNamePrefix, trimmed, time.Now().UnixNano())
}

func resolveCreateNetworkName(network runtimeports.TopologyCreateNetwork) string {
	if name := strings.TrimSpace(network.Name); name != "" {
		return name
	}
	return buildManagedNetworkName(network.Key)
}

func managedContainerLabels(req *runtimeports.TopologyCreateRequest) map[string]string {
	return runtimedomain.ChallengeInstanceLabels(resolveManagedComposeService(req))
}

func resolveManagedComposeService(req *runtimeports.TopologyCreateRequest) string {
	if isAWDTopology(req) {
		return runtimedomain.ComposeServiceAWD
	}
	return runtimedomain.ComposeServiceJeopardy
}

func isAWDTopology(req *runtimeports.TopologyCreateRequest) bool {
	if req == nil {
		return false
	}
	if strings.HasPrefix(strings.TrimSpace(req.ContainerName), awdWorkspaceNamePrefix) {
		return true
	}
	for _, network := range req.Networks {
		if strings.HasPrefix(strings.TrimSpace(network.Name), awdContestNetworkPrefix) {
			return true
		}
	}
	for _, node := range req.Nodes {
		if looksLikeAWDImage(node.Image) {
			return true
		}
		for _, alias := range node.NetworkAliases {
			trimmed := strings.TrimSpace(alias)
			if strings.HasPrefix(trimmed, "awd-c") || strings.HasPrefix(trimmed, "awd-ws-c") {
				return true
			}
		}
	}
	return false
}

func looksLikeAWDImage(image string) bool {
	image = strings.ToLower(strings.TrimSpace(image))
	if image == "" {
		return false
	}
	repository := image
	if digestIndex := strings.Index(repository, "@"); digestIndex >= 0 {
		repository = repository[:digestIndex]
	}
	lastSlash := strings.LastIndex(repository, "/")
	if tagIndex := strings.LastIndex(repository, ":"); tagIndex > lastSlash {
		repository = repository[:tagIndex]
	}
	base := repository
	if lastSlash = strings.LastIndex(repository, "/"); lastSlash >= 0 {
		parent := repository[:lastSlash]
		base = repository[lastSlash+1:]
		if parent == "awd" || strings.HasSuffix(parent, "/awd") {
			return true
		}
	}
	return base == "awd" || strings.HasPrefix(base, "awd-")
}

func normalizedCreateNetworks(networks []runtimeports.TopologyCreateNetwork) []runtimeports.TopologyCreateNetwork {
	if len(networks) == 0 {
		return []runtimeports.TopologyCreateNetwork{{Key: model.TopologyDefaultNetworkKey}}
	}
	return networks
}

func normalizedNodeNetworkKeys(keys []string, networks []runtimeports.TopologyCreateNetwork) []string {
	if len(keys) > 0 {
		return append([]string(nil), keys...)
	}
	return []string{normalizedCreateNetworks(networks)[0].Key}
}

func normalizedNetworkAliases(aliases []string) []string {
	if len(aliases) == 0 {
		return nil
	}
	result := make([]string, 0, len(aliases))
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		trimmed := strings.TrimSpace(alias)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func collectOwnedNetworkIDs(networks []createdTopologyNetwork) []string {
	result := make([]string, 0, len(networks))
	for _, network := range networks {
		if network.id != "" && !network.shared {
			result = append(result, network.id)
		}
	}
	return result
}
