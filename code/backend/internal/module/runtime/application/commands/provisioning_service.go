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
)

type provisioningRepository interface {
	ListAllocatedPorts() ([]int, error)
}

type provisioningEngine interface {
	CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool) (string, error)
	CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error)
	ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error)
	ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error
	InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	ApplyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
}

type createdTopologyNetwork struct {
	key      string
	name     string
	id       string
	internal bool
}

// ProvisioningService 收口运行时资源创建编排，包括单容器与拓扑实例创建。
type ProvisioningService struct {
	repo   provisioningRepository
	engine provisioningEngine
	config *config.ContainerConfig
	logger *zap.Logger
}

// NewProvisioningService 创建运行时资源编排服务。
func NewProvisioningService(repo provisioningRepository, engine provisioningEngine, cfg *config.ContainerConfig, logger *zap.Logger) *ProvisioningService {
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
				Key:          "default",
				Image:        imageName,
				Env:          env,
				ServicePort:  servicePort,
				IsEntryPoint: true,
				NetworkKeys:  []string{model.TopologyDefaultNetworkKey},
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

	hostPort := req.ReservedHostPort
	if hostPort <= 0 {
		var err error
		hostPort, err = s.allocatePort()
		if err != nil {
			return nil, err
		}
	}

	createdNetworks := make([]createdTopologyNetwork, 0, len(networks))
	networkByKey := make(map[string]createdTopologyNetwork, len(networks))
	for _, network := range networks {
		networkName := buildManagedNetworkName(network.Key)
		networkID, err := s.engine.CreateNetwork(ctx, networkName, managedNetworkLabels(), network.Internal)
		if err != nil {
			s.cleanupTopologyResources(ctx, nil, collectCreatedNetworkIDs(createdNetworks))
			return nil, err
		}
		item := createdTopologyNetwork{
			key:      network.Key,
			name:     networkName,
			id:       networkID,
			internal: network.Internal,
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
		})
	}

	createdContainerIDs := make([]string, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodeNetworkKeys := normalizedNodeNetworkKeys(node.NetworkKeys, networks)
		primaryNetwork := networkByKey[nodeNetworkKeys[0]]
		ports := map[string]string(nil)
		if node.IsEntryPoint {
			ports = map[string]string{
				strconv.Itoa(node.ServicePort): strconv.Itoa(hostPort),
			}
		}

		containerID, err := s.engine.CreateContainer(ctx, &model.ContainerConfig{
			Image:     node.Image,
			Name:      buildManagedContainerName(),
			Env:       envMapToList(node.Env),
			Ports:     ports,
			Labels:    managedContainerLabels(),
			Resources: node.Resources,
			Network:   primaryNetwork.name,
		})
		if err != nil {
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
			return nil, err
		}
		if err := s.engine.StartContainer(ctx, containerID); err != nil {
			createdContainerIDs = append(createdContainerIDs, containerID)
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
			return nil, err
		}
		for _, networkKey := range nodeNetworkKeys[1:] {
			if err := s.engine.ConnectContainerToNetwork(ctx, containerID, networkByKey[networkKey].name); err != nil {
				createdContainerIDs = append(createdContainerIDs, containerID)
				s.cleanupTopologyResources(ctx, createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
				return nil, err
			}
		}

		createdContainerIDs = append(createdContainerIDs, containerID)
		runtimeItem := model.InstanceRuntimeContainer{
			NodeKey:      node.Key,
			ContainerID:  containerID,
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.IsEntryPoint,
			NetworkKeys:  append([]string(nil), nodeNetworkKeys...),
		}
		if node.IsEntryPoint {
			runtimeItem.HostPort = hostPort
		}
		details.Containers = append(details.Containers, runtimeItem)
	}

	resolvedACLRules, err := s.resolveTopologyACLRules(ctx, req, details)
	if err != nil {
		s.cleanupTopologyResources(ctx, createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
		return nil, err
	}
	if len(resolvedACLRules) > 0 {
		if err := s.engine.ApplyACLRules(ctx, resolvedACLRules); err != nil {
			s.cleanupTopologyResources(ctx, createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
			return nil, err
		}
		details.ACLRules = resolvedACLRules
	}

	return &runtimeports.TopologyCreateResult{
		PrimaryContainerID: details.Containers[entryNodeIndex].ContainerID,
		NetworkID:          details.Networks[0].NetworkID,
		AccessURL:          fmt.Sprintf("http://%s:%d", s.config.PublicHost, hostPort),
		RuntimeDetails:     details,
	}, nil
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

func (s *ProvisioningService) allocatePort() (int, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("runtime provisioning repository is not configured")
	}

	usedPorts, err := s.repo.ListAllocatedPorts()
	if err != nil {
		return 0, err
	}

	used := make(map[int]struct{}, len(usedPorts))
	for _, port := range usedPorts {
		used[port] = struct{}{}
	}

	for port := s.config.PortRangeStart; port < s.config.PortRangeEnd; port++ {
		if _, exists := used[port]; exists {
			continue
		}
		return port, nil
	}
	return 0, fmt.Errorf("no available port in range %d-%d", s.config.PortRangeStart, s.config.PortRangeEnd)
}

func (s *ProvisioningService) cleanupTopologyResources(ctx context.Context, containerIDs []string, networkIDs []string) {
	for idx := len(containerIDs) - 1; idx >= 0; idx-- {
		_ = s.removeContainerWithContext(ctx, containerIDs[idx])
	}
	for idx := len(networkIDs) - 1; idx >= 0; idx-- {
		_ = s.removeNetworkWithContext(ctx, networkIDs[idx])
	}
}

func (s *ProvisioningService) removeContainerWithContext(ctx context.Context, containerID string) error {
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

func (s *ProvisioningService) removeNetworkWithContext(ctx context.Context, networkID string) error {
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

func buildManagedContainerName() string {
	return fmt.Sprintf("%s%d", managedContainerNamePrefix, time.Now().UnixNano())
}

func buildManagedNetworkName(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		trimmed = model.TopologyDefaultNetworkKey
	}
	return fmt.Sprintf("%s%s-%d", managedNetworkNamePrefix, trimmed, time.Now().UnixNano())
}

func managedContainerLabels() map[string]string {
	return map[string]string{
		runtimedomain.ManagedByLabelKey:         runtimedomain.ManagedByLabelValue,
		runtimedomain.ChallengeInstanceLabelKey: runtimedomain.ChallengeInstanceLabelValue,
	}
}

func managedNetworkLabels() map[string]string {
	return map[string]string{
		runtimedomain.ManagedByLabelKey:         runtimedomain.ManagedByLabelValue,
		runtimedomain.ChallengeInstanceLabelKey: runtimedomain.ChallengeInstanceLabelValue,
	}
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

func collectCreatedNetworkIDs(networks []createdTopologyNetwork) []string {
	result := make([]string, 0, len(networks))
	for _, network := range networks {
		if network.id != "" {
			result = append(result, network.id)
		}
	}
	return result
}
