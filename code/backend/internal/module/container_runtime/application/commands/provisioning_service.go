package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimedomain "ctf-platform/internal/module/container_runtime/domain"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	"ctf-platform/internal/platform/logctx"
)

const (
	managedContainerNamePrefix = "ctf-instance-"
	managedNetworkNamePrefix   = "ctf-net-"
	awdContestNetworkPrefix    = "ctf-awd-contest-"
	awdWorkspaceNamePrefix     = "ctf-workspace-"
)

type ProvisioningRepository interface {
	ReserveAvailablePort(ctx context.Context, start, end int) (int, error)
	ReleaseReservedPort(ctx context.Context, port int) error
	ReserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int) (string, error)
	ReserveAvailableSubnetForInstance(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string) (string, error)
	ReserveAvailableSubnetExcluding(ctx context.Context, baseCIDR string, subnetMask int, excludedSubnets []string) (string, error)
	ReserveAvailableSubnetForInstanceExcluding(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error)
	ReleaseReservedSubnet(ctx context.Context, subnet string) error
	ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error
}

type nodeScopedSubnetRepository interface {
	ReserveAvailableSubnetForNode(ctx context.Context, nodeID int64, poolKind string, instanceID int64, networkKey string) (string, error)
	QuarantineSubnet(ctx context.Context, nodeID int64, subnet string, reason string) error
}

type createdTopologyNetwork struct {
	key      string
	name     string
	id       string
	subnet   string
	internal bool
	shared   bool
}

type topologyStageContext struct {
	instanceID  int64
	stage       string
	nodeKey     string
	image       string
	networkKey  string
	networkName string
	subnet      string
	hostPort    int
	containerID string
}

// ProvisioningService 收口运行时资源创建编排，包括单容器与拓扑实例创建。
type ProvisioningService struct {
	repo   ProvisioningRepository
	engine runtimeports.ContainerProvisioningRuntime
	config *config.ContainerConfig
	logger *zap.Logger
}

// NewProvisioningService 创建运行时资源编排服务。
func NewProvisioningService(repo ProvisioningRepository, engine runtimeports.ContainerProvisioningRuntime, cfg *config.ContainerConfig, logger *zap.Logger) *ProvisioningService {
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
	applyProvisioningConfigDefaults(cfg)
	return &ProvisioningService{
		repo:   repo,
		engine: engine,
		config: cfg,
		logger: logger,
	}
}

func applyProvisioningConfigDefaults(cfg *config.ContainerConfig) {
	if cfg == nil {
		return
	}
	if strings.TrimSpace(cfg.Network.SingleContainerSubnetBase) == "" {
		cfg.Network.SingleContainerSubnetBase = "10.11.0.0/16"
	}
	if cfg.Network.SingleContainerSubnetMask <= 0 {
		cfg.Network.SingleContainerSubnetMask = 29
	}
	if strings.TrimSpace(cfg.Network.TopologySubnetBase) == "" {
		cfg.Network.TopologySubnetBase = "10.10.0.0/16"
	}
	if cfg.Network.TopologySubnetMask <= 0 {
		cfg.Network.TopologySubnetMask = 24
	}
}

// CreateContainer 为单容器题目创建默认拓扑，并返回入口容器与端口信息。
func (s *ProvisioningService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	servicePort, err = s.resolveServicePort(ctx, imageName)
	if err != nil {
		return "", "", 0, 0, err
	}

	result, err := s.CreateTopology(ctx, &runtimecontracts.TopologyCreateRequest{
		SubnetPool:       runtimecontracts.SubnetPoolSingleContainer,
		ReservedHostPort: reservedHostPort,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
			{
				Key:             "default",
				Image:           imageName,
				Env:             env,
				ServicePort:     servicePort,
				ServiceProtocol: runtimecontracts.ChallengeTargetProtocolHTTP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
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
func (s *ProvisioningService) CreateTopology(ctx context.Context, req *runtimecontracts.TopologyCreateRequest) (*runtimecontracts.TopologyCreateResult, error) {
	ctx = normalizeContext(ctx)
	if req == nil || len(req.Nodes) == 0 {
		return nil, fmt.Errorf("topology nodes are required")
	}
	if s.engine == nil {
		return nil, runtimeports.ErrRuntimeEngineUnavailable
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
	occupiedSubnets := make([]string, 0, len(networks))
	var err error
	if topologyNeedsRuntimeOccupiedSubnets(networks) {
		occupiedSubnets, err = s.listRuntimeOccupiedSubnets(ctx)
		if err != nil {
			return nil, err
		}
	}
	for _, network := range networks {
		networkName := resolveCreateNetworkName(network)
		var (
			subnet    string
			networkID string
		)
		for {
			subnet, err = s.allocateNetworkSubnet(ctx, req, req.OwnerInstanceID, network, occupiedSubnets)
			if err != nil {
				s.cleanupTopologyResources(ctx, nil, createdNetworks, req.OwnerInstanceID)
				return nil, err
			}
			occupiedSubnets = appendUniqueSubnet(occupiedSubnets, subnet)
			finish := s.startTopologyStage(ctx)
			networkID, err = s.engine.CreateNetwork(ctx, networkName, managedLabels, network.Internal, network.Shared, subnet)
			finish(err, topologyStageContext{
				instanceID:  req.OwnerInstanceID,
				stage:       "network_create",
				networkKey:  network.Key,
				networkName: networkName,
				subnet:      subnet,
			})
			if err == nil {
				break
			}
			if errors.Is(err, runtimeports.ErrRuntimeNetworkSubnetConflict) && canRetrySubnetAllocation(network, subnet) {
				s.quarantineOrReleaseConflictedNetworkSubnet(ctx, req, subnet, err)
				occupiedSubnets = appendUniqueSubnet(occupiedSubnets, subnet)
				continue
			}
			s.releaseNetworkSubnet(ctx, req.OwnerInstanceID, subnet)
			s.cleanupTopologyResources(ctx, nil, createdNetworks, req.OwnerInstanceID)
			return nil, err
		}
		item := createdTopologyNetwork{
			key:      network.Key,
			name:     networkName,
			id:       networkID,
			subnet:   subnet,
			internal: network.Internal,
			shared:   network.Shared,
		}
		createdNetworks = append(createdNetworks, item)
		networkByKey[network.Key] = item
	}

	details := runtimecontracts.InstanceRuntimeDetails{
		Networks:   make([]runtimecontracts.InstanceRuntimeNetwork, 0, len(createdNetworks)),
		Containers: make([]runtimecontracts.InstanceRuntimeContainer, 0, len(req.Nodes)),
	}
	for _, network := range createdNetworks {
		details.Networks = append(details.Networks, runtimecontracts.InstanceRuntimeNetwork{
			Key:       network.key,
			Name:      network.name,
			NetworkID: network.id,
			Subnet:    network.subnet,
			Internal:  network.internal,
			Shared:    network.shared,
		})
	}

	createdContainerIDs := make([]string, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodeNetworkKeys := normalizedNodeNetworkKeys(node.NetworkKeys, networks)
		primaryNetwork := networkByKey[nodeNetworkKeys[0]]
		nodeHostPort := 0
		if node.IsEntryPoint && publishEntryPort {
			nodeHostPort = hostPort
		}
		servicePort := node.ServicePort
		if node.IsEntryPoint && servicePort <= 0 {
			finish := s.startTopologyStage(ctx)
			resolvedPort, err := s.resolveServicePort(ctx, node.Image)
			finish(err, topologyStageContext{
				instanceID:  req.OwnerInstanceID,
				stage:       "service_port_resolve",
				nodeKey:     node.Key,
				image:       node.Image,
				networkKey:  primaryNetwork.key,
				networkName: primaryNetwork.name,
				hostPort:    nodeHostPort,
			})
			if err != nil {
				s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
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

		finish := s.startTopologyStage(ctx)
		containerID, err := s.engine.CreateContainer(ctx, &runtimecontracts.ContainerConfig{
			Image:          node.Image,
			Name:           buildManagedContainerName(req.ContainerName),
			Env:            envMapToList(node.Env),
			Command:        append([]string(nil), node.Command...),
			WorkingDir:     strings.TrimSpace(node.WorkingDir),
			Ports:          ports,
			Mounts:         append([]runtimecontracts.ContainerMount(nil), node.Mounts...),
			Labels:         managedLabels,
			Resources:      node.Resources,
			Network:        primaryNetwork.name,
			NetworkAliases: normalizedNetworkAliases(node.NetworkAliases),
		})
		finish(err, topologyStageContext{
			instanceID:  req.OwnerInstanceID,
			stage:       "container_create",
			nodeKey:     node.Key,
			image:       node.Image,
			networkKey:  primaryNetwork.key,
			networkName: primaryNetwork.name,
			subnet:      primaryNetwork.subnet,
			hostPort:    nodeHostPort,
			containerID: containerID,
		})
		if err != nil {
			s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
			return nil, err
		}
		finish = s.startTopologyStage(ctx)
		err = s.engine.StartContainer(ctx, containerID)
		finish(err, topologyStageContext{
			instanceID:  req.OwnerInstanceID,
			stage:       "container_start",
			nodeKey:     node.Key,
			image:       node.Image,
			networkKey:  primaryNetwork.key,
			networkName: primaryNetwork.name,
			subnet:      primaryNetwork.subnet,
			hostPort:    nodeHostPort,
			containerID: containerID,
		})
		if err != nil {
			createdContainerIDs = append(createdContainerIDs, containerID)
			s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
			return nil, err
		}
		for _, networkKey := range nodeNetworkKeys[1:] {
			extraNetwork := networkByKey[networkKey]
			finish = s.startTopologyStage(ctx)
			err = s.engine.ConnectContainerToNetwork(ctx, containerID, extraNetwork.name)
			finish(err, topologyStageContext{
				instanceID:  req.OwnerInstanceID,
				stage:       "connect_extra_networks",
				nodeKey:     node.Key,
				image:       node.Image,
				networkKey:  extraNetwork.key,
				networkName: extraNetwork.name,
				subnet:      extraNetwork.subnet,
				hostPort:    nodeHostPort,
				containerID: containerID,
			})
			if err != nil {
				createdContainerIDs = append(createdContainerIDs, containerID)
				s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
				return nil, err
			}
		}
		finish = s.startTopologyStage(ctx)
		networkIPs, err := s.engine.InspectContainerNetworkIPs(ctx, containerID)
		finish(err, topologyStageContext{
			instanceID:  req.OwnerInstanceID,
			stage:       "inspect_network_ips",
			nodeKey:     node.Key,
			image:       node.Image,
			networkKey:  primaryNetwork.key,
			networkName: primaryNetwork.name,
			subnet:      primaryNetwork.subnet,
			hostPort:    nodeHostPort,
			containerID: containerID,
		})
		if err != nil {
			createdContainerIDs = append(createdContainerIDs, containerID)
			s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
			return nil, err
		}

		createdContainerIDs = append(createdContainerIDs, containerID)
		serviceProtocol := normalizeServiceProtocol(node.ServiceProtocol)
		runtimeItem := runtimecontracts.InstanceRuntimeContainer{
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
		s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
		return nil, err
	}
	if len(resolvedACLRules) > 0 {
		if req.OwnerInstanceID <= 0 {
			s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
			return nil, fmt.Errorf("topology acl requires a stable owner instance id, got %d", req.OwnerInstanceID)
		}
		handle := &runtimecontracts.InstanceRuntimeACLHandle{
			Chain: fmt.Sprintf("CTF-INS-%d", req.OwnerInstanceID),
		}
		if err := s.engine.ApplyACL(ctx, handle, resolvedACLRules); err != nil {
			s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
			return nil, err
		}
		details.ACL = handle
		details.ACLRules = resolvedACLRules
	}

	accessURL, err := s.resolveEntryAccessURL(ctx, details, entryNodeIndex, publishEntryPort, hostPort)
	if err != nil {
		s.cleanupTopologyResources(ctx, createdContainerIDs, createdNetworks, req.OwnerInstanceID)
		return nil, err
	}

	success = true
	return &runtimecontracts.TopologyCreateResult{
		PrimaryContainerID: details.Containers[entryNodeIndex].ContainerID,
		NetworkID:          details.Networks[0].NetworkID,
		AccessURL:          accessURL,
		RuntimeDetails:     details,
	}, nil
}

func (s *ProvisioningService) resolveEntryAccessURL(ctx context.Context, details runtimecontracts.InstanceRuntimeDetails, entryNodeIndex int, publishEntryPort bool, hostPort int) (string, error) {
	if entryNodeIndex < 0 || entryNodeIndex >= len(details.Containers) {
		return "", fmt.Errorf("entry container is missing")
	}
	entry := details.Containers[entryNodeIndex]
	scheme := normalizeServiceProtocol(entry.ServiceProtocol)
	if publishEntryPort {
		host := runtimecontracts.ResolveRuntimePublishedAccessHost(s.config.PublicHost, s.config.AccessHost)
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
	case runtimecontracts.ChallengeTargetProtocolTCP:
		return runtimecontracts.ChallengeTargetProtocolTCP
	default:
		return runtimecontracts.ChallengeTargetProtocolHTTP
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

func (s *ProvisioningService) resolveTopologyACLRules(ctx context.Context, req *runtimecontracts.TopologyCreateRequest, details runtimecontracts.InstanceRuntimeDetails) ([]runtimecontracts.InstanceRuntimeACLRule, error) {
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

func (s *ProvisioningService) allocateNetworkSubnet(ctx context.Context, req *runtimecontracts.TopologyCreateRequest, ownerInstanceID int64, network runtimecontracts.TopologyCreateNetwork, excludedSubnets []string) (string, error) {
	if network.Shared {
		return "", nil
	}
	if subnet := strings.TrimSpace(network.Subnet); subnet != "" {
		return subnet, nil
	}
	if s.repo == nil {
		return "", fmt.Errorf("runtime provisioning repository is not configured")
	}
	if nodeID := runtimeNodeIDValue(req); nodeID > 0 {
		if nodeRepo, ok := s.repo.(nodeScopedSubnetRepository); ok {
			return nodeRepo.ReserveAvailableSubnetForNode(ctx, nodeID, string(resolveSubnetPoolKind(req)), ownerInstanceID, network.Key)
		}
	}
	baseCIDR, subnetMask := s.resolveNetworkPool(req)
	if ownerInstanceID > 0 {
		return s.repo.ReserveAvailableSubnetForInstanceExcluding(ctx, baseCIDR, subnetMask, ownerInstanceID, network.Key, excludedSubnets)
	}
	return s.repo.ReserveAvailableSubnetExcluding(ctx, baseCIDR, subnetMask, excludedSubnets)
}

func (s *ProvisioningService) quarantineOrReleaseConflictedNetworkSubnet(ctx context.Context, req *runtimecontracts.TopologyCreateRequest, subnet string, cause error) {
	if s == nil || s.repo == nil {
		return
	}
	if nodeID := runtimeNodeIDValue(req); nodeID > 0 {
		if nodeRepo, ok := s.repo.(nodeScopedSubnetRepository); ok {
			_ = nodeRepo.QuarantineSubnet(ctx, nodeID, subnet, cause.Error())
			return
		}
	}
	s.releaseNetworkSubnet(ctx, req.OwnerInstanceID, subnet)
}

func runtimeNodeIDValue(req *runtimecontracts.TopologyCreateRequest) int64 {
	if req == nil {
		return 0
	}
	return req.RuntimeNodeID
}

func (s *ProvisioningService) resolveNetworkPool(req *runtimecontracts.TopologyCreateRequest) (string, int) {
	switch resolveSubnetPoolKind(req) {
	case runtimecontracts.SubnetPoolSingleContainer:
		return s.config.Network.SingleContainerSubnetBase, s.config.Network.SingleContainerSubnetMask
	default:
		return s.config.Network.TopologySubnetBase, s.config.Network.TopologySubnetMask
	}
}

func resolveSubnetPoolKind(req *runtimecontracts.TopologyCreateRequest) runtimecontracts.SubnetPoolKind {
	if req == nil {
		return runtimecontracts.SubnetPoolTopology
	}
	switch req.SubnetPool {
	case runtimecontracts.SubnetPoolSingleContainer:
		return runtimecontracts.SubnetPoolSingleContainer
	default:
		return runtimecontracts.SubnetPoolTopology
	}
}

func (s *ProvisioningService) releaseNetworkSubnet(ctx context.Context, ownerInstanceID int64, subnet string) {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" || s.repo == nil {
		return
	}
	if ownerInstanceID > 0 {
		_ = s.repo.ReleaseSubnetForInstance(ctx, subnet, ownerInstanceID)
		return
	}
	_ = s.repo.ReleaseReservedSubnet(ctx, subnet)
}

func (s *ProvisioningService) cleanupTopologyResources(ctx context.Context, containerIDs []string, networks []createdTopologyNetwork, ownerInstanceID int64) {
	for idx := len(containerIDs) - 1; idx >= 0; idx-- {
		_ = s.removeContainer(ctx, containerIDs[idx])
	}
	for idx := len(networks) - 1; idx >= 0; idx-- {
		if !networks[idx].shared {
			_ = s.removeNetwork(ctx, networks[idx].id)
			s.releaseNetworkSubnet(ctx, ownerInstanceID, networks[idx].subnet)
		}
	}
}

func (s *ProvisioningService) removeContainer(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}
	if s.engine == nil {
		logctx.Info(ctx, s.logger, "删除容器（降级模拟）", zap.String("container_id", containerID))
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(normalizeContext(ctx), 10*time.Second)
	defer cancel()
	_ = s.engine.StopContainer(timeoutCtx, containerID, 5*time.Second)
	if err := s.engine.RemoveContainer(timeoutCtx, containerID, true); err != nil {
		return err
	}

	logctx.Info(ctx, s.logger, "删除容器", zap.String("container_id", containerID))
	return nil
}

func (s *ProvisioningService) removeNetwork(ctx context.Context, networkID string) error {
	if networkID == "" {
		return nil
	}
	if s.engine == nil {
		logctx.Info(ctx, s.logger, "删除网络（降级跳过）", zap.String("network_id", networkID))
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(normalizeContext(ctx), 10*time.Second)
	defer cancel()
	if err := s.engine.RemoveNetwork(timeoutCtx, networkID); err != nil {
		return err
	}

	logctx.Info(ctx, s.logger, "删除网络", zap.String("network_id", networkID))
	return nil
}

func (s *ProvisioningService) startTopologyStage(ctx context.Context) func(error, topologyStageContext) {
	startedAt := time.Now()
	return func(err error, stageCtx topologyStageContext) {
		s.logTopologyStage(ctx, time.Since(startedAt), err, stageCtx)
	}
}

func (s *ProvisioningService) logTopologyStage(ctx context.Context, duration time.Duration, err error, stageCtx topologyStageContext) {
	fields := make([]zap.Field, 0, 10)
	fields = append(fields,
		zap.String("stage", stageCtx.stage),
		zap.Duration("duration", duration),
	)
	if stageCtx.instanceID > 0 {
		fields = append(fields, zap.Int64("instance_id", stageCtx.instanceID))
	}
	if stageCtx.nodeKey != "" {
		fields = append(fields, zap.String("node_key", stageCtx.nodeKey))
	}
	if stageCtx.image != "" {
		fields = append(fields, zap.String("image", stageCtx.image))
	}
	if stageCtx.networkKey != "" {
		fields = append(fields, zap.String("network_key", stageCtx.networkKey))
	}
	if stageCtx.networkName != "" {
		fields = append(fields, zap.String("network_name", stageCtx.networkName))
	}
	if stageCtx.subnet != "" {
		fields = append(fields, zap.String("subnet", stageCtx.subnet))
	}
	if stageCtx.hostPort > 0 {
		fields = append(fields, zap.Int("host_port", stageCtx.hostPort))
	}
	if stageCtx.containerID != "" {
		fields = append(fields, zap.String("container_id", stageCtx.containerID))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		logctx.Warn(ctx, s.logger, "runtime provisioning stage failed", fields...)
		return
	}
	logctx.Info(ctx, s.logger, "runtime provisioning stage succeeded", fields...)
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
		trimmed = runtimecontracts.TopologyDefaultNetworkKey
	}
	return fmt.Sprintf("%s%s-%d", managedNetworkNamePrefix, trimmed, time.Now().UnixNano())
}

func resolveCreateNetworkName(network runtimecontracts.TopologyCreateNetwork) string {
	if name := strings.TrimSpace(network.Name); name != "" {
		return name
	}
	return buildManagedNetworkName(network.Key)
}

func canRetrySubnetAllocation(network runtimecontracts.TopologyCreateNetwork, subnet string) bool {
	return strings.TrimSpace(subnet) != "" && !network.Shared && strings.TrimSpace(network.Subnet) == ""
}

func (s *ProvisioningService) listRuntimeOccupiedSubnets(ctx context.Context) ([]string, error) {
	if s == nil || s.engine == nil {
		return nil, nil
	}

	subnets, err := s.engine.ListNetworkSubnets(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(subnets))
	for _, subnet := range subnets {
		result = appendUniqueSubnet(result, subnet)
	}
	return result, nil
}

func appendUniqueSubnet(items []string, subnet string) []string {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		return items
	}
	for _, item := range items {
		if item == subnet {
			return items
		}
	}
	return append(items, subnet)
}

func managedContainerLabels(req *runtimecontracts.TopologyCreateRequest) map[string]string {
	return runtimecontracts.ChallengeInstanceLabels(resolveManagedComposeService(req))
}

func resolveManagedComposeService(req *runtimecontracts.TopologyCreateRequest) string {
	if isAWDTopology(req) {
		return runtimecontracts.ComposeServiceAWD
	}
	return runtimecontracts.ComposeServiceJeopardy
}

func isAWDTopology(req *runtimecontracts.TopologyCreateRequest) bool {
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

func normalizedCreateNetworks(networks []runtimecontracts.TopologyCreateNetwork) []runtimecontracts.TopologyCreateNetwork {
	if len(networks) == 0 {
		return []runtimecontracts.TopologyCreateNetwork{{Key: runtimecontracts.TopologyDefaultNetworkKey}}
	}
	return networks
}

func topologyNeedsRuntimeOccupiedSubnets(networks []runtimecontracts.TopologyCreateNetwork) bool {
	for _, network := range normalizedCreateNetworks(networks) {
		if network.Shared {
			continue
		}
		if strings.TrimSpace(network.Subnet) != "" {
			continue
		}
		return true
	}
	return false
}

func normalizedNodeNetworkKeys(keys []string, networks []runtimecontracts.TopologyCreateNetwork) []string {
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
