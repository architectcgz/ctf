package runtime

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Service struct {
	repo   *Repository
	engine runtimeEngine
	config *config.ContainerConfig
	logger *zap.Logger
}

type TopologyCreateNode struct {
	Key          string
	Image        string
	Env          map[string]string
	ServicePort  int
	IsEntryPoint bool
	NetworkKeys  []string
	Resources    *model.ResourceLimits
}

type TopologyCreateNetwork struct {
	Key      string
	Internal bool
}

type TopologyCreateRequest struct {
	Networks         []TopologyCreateNetwork
	Nodes            []TopologyCreateNode
	Policies         []model.TopologyTrafficPolicy
	ReservedHostPort int
}

type TopologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}

type createdTopologyNetwork struct {
	key      string
	name     string
	id       string
	internal bool
}

const (
	managedByLabelKey           = "managed-by"
	managedByLabelValue         = "ctf-platform"
	challengeInstanceLabelKey   = "ctf-component"
	challengeInstanceLabelValue = "challenge-instance"
	managedContainerNamePrefix  = "ctf-instance-"
	managedNetworkNamePrefix    = "ctf-net-"
)

type runtimeEngine interface {
	CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool) (string, error)
	CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error)
	ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error)
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
	ListManagedContainerStats(ctx context.Context, managedBy string) ([]ManagedContainerStat, error)
	ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error
	InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	ApplyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
	RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ListManagedContainers(ctx context.Context, managedBy string) ([]ManagedContainer, error)
}

func NewService(repo *Repository, engine runtimeEngine, cfg *config.ContainerConfig, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	if isNilRuntimeEngine(engine) {
		engine = nil
	}
	return &Service{
		repo:   repo,
		engine: engine,
		config: cfg,
		logger: logger,
	}
}

func isNilRuntimeEngine(engine runtimeEngine) bool {
	if engine == nil {
		return true
	}
	value := reflect.ValueOf(engine)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (s *Service) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	servicePort, err = s.resolveServicePort(ctx, imageName)
	if err != nil {
		return "", "", 0, 0, err
	}

	result, err := s.CreateTopology(ctx, &TopologyCreateRequest{
		ReservedHostPort: reservedHostPort,
		Networks: []TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []TopologyCreateNode{
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
	hostPort = 0
	for _, item := range result.RuntimeDetails.Containers {
		if item.IsEntryPoint {
			hostPort = item.HostPort
			break
		}
	}
	return result.PrimaryContainerID, result.NetworkID, hostPort, servicePort, nil
}

func (s *Service) resolveServicePort(ctx context.Context, imageRef string) (int, error) {
	preferredPort := s.config.DefaultExposedPort
	if preferredPort <= 0 {
		preferredPort = 8080
	}
	if s.engine == nil {
		return preferredPort, nil
	}

	resolvedPort, err := s.engine.ResolveServicePort(ctx, imageRef, preferredPort)
	if err != nil {
		return 0, err
	}
	if resolvedPort <= 0 {
		return preferredPort, nil
	}
	return resolvedPort, nil
}

func (s *Service) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	if strings.TrimSpace(imageRef) == "" || s.engine == nil {
		return 0, nil
	}
	return s.engine.InspectImageSize(ctx, imageRef)
}

func (s *Service) RemoveImage(ctx context.Context, imageRef string) error {
	if strings.TrimSpace(imageRef) == "" || s.engine == nil {
		return nil
	}
	return s.engine.RemoveImage(ctx, imageRef)
}

func (s *Service) ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error) {
	if s.engine == nil {
		return []ManagedContainerStat{}, nil
	}
	return s.engine.ListManagedContainerStats(ctx, managedByFilter())
}

func (s *Service) CleanupRuntime(instance *model.Instance) error {
	return s.CleanupRuntimeWithContext(context.Background(), instance)
}

func (s *Service) CleanupRuntimeWithContext(ctx context.Context, instance *model.Instance) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if instance == nil {
		return nil
	}
	if err := s.removeACLRulesWithContext(ctx, managedACLRules(instance)); err != nil {
		s.logger.Warn("删除实例 ACL 规则失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	for _, containerID := range managedContainerIDs(instance) {
		if err := s.RemoveContainerWithContext(ctx, containerID); err != nil {
			return err
		}
	}
	for _, networkID := range managedNetworkIDs(instance) {
		if err := s.RemoveNetworkWithContext(ctx, networkID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error) {
	if req == nil || len(req.Nodes) == 0 {
		return nil, fmt.Errorf("topology nodes are required")
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

	if s.engine == nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}

		details := model.InstanceRuntimeDetails{
			Networks:   make([]model.InstanceRuntimeNetwork, 0, len(networks)),
			Containers: make([]model.InstanceRuntimeContainer, 0, len(req.Nodes)),
		}
		for _, network := range networks {
			details.Networks = append(details.Networks, model.InstanceRuntimeNetwork{
				Key:       network.Key,
				Name:      network.Key,
				NetworkID: fmt.Sprintf("net-%d-%s", time.Now().UnixNano(), network.Key),
				Internal:  network.Internal,
			})
		}
		for idx, node := range req.Nodes {
			containerID := fmt.Sprintf("ctf-%d-%d", time.Now().UnixNano(), idx)
			item := model.InstanceRuntimeContainer{
				NodeKey:      node.Key,
				ContainerID:  containerID,
				ServicePort:  node.ServicePort,
				IsEntryPoint: node.IsEntryPoint,
				NetworkKeys:  append([]string(nil), normalizedNodeNetworkKeys(node.NetworkKeys, networks)...),
			}
			if node.IsEntryPoint {
				item.HostPort = hostPort
			}
			details.Containers = append(details.Containers, item)
		}
		return &TopologyCreateResult{
			PrimaryContainerID: details.Containers[entryNodeIndex].ContainerID,
			NetworkID:          details.Networks[0].NetworkID,
			AccessURL:          fmt.Sprintf("http://%s:%d", s.config.PublicHost, hostPort),
			RuntimeDetails:     details,
		}, nil
	}

	createdNetworks := make([]createdTopologyNetwork, 0, len(networks))
	networkByKey := make(map[string]createdTopologyNetwork, len(networks))
	for _, network := range networks {
		networkName := buildManagedNetworkName(network.Key)
		networkID, createErr := s.engine.CreateNetwork(ctx, networkName, managedNetworkLabels(), network.Internal)
		if createErr != nil {
			s.cleanupTopologyResources(nil, collectCreatedNetworkIDs(createdNetworks))
			return nil, createErr
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
		containerID, createErr := s.engine.CreateContainer(ctx, &model.ContainerConfig{
			Image:     node.Image,
			Name:      buildManagedContainerName(),
			Env:       envMapToList(node.Env),
			Ports:     ports,
			Labels:    managedContainerLabels(),
			Resources: node.Resources,
			Network:   primaryNetwork.name,
		})
		if createErr != nil {
			s.cleanupTopologyResources(createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
			return nil, createErr
		}
		if startErr := s.engine.StartContainer(ctx, containerID); startErr != nil {
			createdContainerIDs = append(createdContainerIDs, containerID)
			s.cleanupTopologyResources(createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
			return nil, startErr
		}
		for _, networkKey := range nodeNetworkKeys[1:] {
			if connectErr := s.engine.ConnectContainerToNetwork(ctx, containerID, networkByKey[networkKey].name); connectErr != nil {
				createdContainerIDs = append(createdContainerIDs, containerID)
				s.cleanupTopologyResources(createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
				return nil, connectErr
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
		s.cleanupTopologyResources(createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
		return nil, err
	}
	if len(resolvedACLRules) > 0 {
		if err := s.engine.ApplyACLRules(ctx, resolvedACLRules); err != nil {
			s.cleanupTopologyResources(createdContainerIDs, collectCreatedNetworkIDs(createdNetworks))
			return nil, err
		}
		details.ACLRules = resolvedACLRules
	}

	return &TopologyCreateResult{
		PrimaryContainerID: details.Containers[entryNodeIndex].ContainerID,
		NetworkID:          details.Networks[0].NetworkID,
		AccessURL:          fmt.Sprintf("http://%s:%d", s.config.PublicHost, hostPort),
		RuntimeDetails:     details,
	}, nil
}

// RemoveContainer 删除容器
func (s *Service) RemoveContainer(containerID string) error {
	return s.RemoveContainerWithContext(context.Background(), containerID)
}

func (s *Service) RemoveContainerWithContext(ctx context.Context, containerID string) error {
	if s.engine == nil {
		s.logger.Info("删除容器（降级模拟）", zap.String("container_id", containerID))
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = s.engine.StopContainer(ctx, containerID, 5*time.Second)
	if err := s.engine.RemoveContainer(ctx, containerID, true); err != nil {
		return err
	}

	s.logger.Info("删除容器", zap.String("container_id", containerID))
	return nil
}

func (s *Service) removeACLRules(rules []model.InstanceRuntimeACLRule) error {
	return s.removeACLRulesWithContext(context.Background(), rules)
}

func (s *Service) removeACLRulesWithContext(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	if len(rules) == 0 || s.engine == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.engine.RemoveACLRules(ctx, rules)
}

// RemoveNetwork 删除网络
func (s *Service) RemoveNetwork(networkID string) error {
	return s.RemoveNetworkWithContext(context.Background(), networkID)
}

func (s *Service) RemoveNetworkWithContext(ctx context.Context, networkID string) error {
	if networkID == "" {
		return nil
	}
	if s.engine == nil {
		s.logger.Info("删除网络（降级跳过）", zap.String("network_id", networkID))
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.engine.RemoveNetwork(ctx, networkID); err != nil {
		return err
	}

	s.logger.Info("删除网络", zap.String("network_id", networkID))
	return nil
}

func (s *Service) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	if s.engine == nil {
		s.logger.Info("写入容器文件（降级跳过）", zap.String("container_id", containerID), zap.String("path", filePath))
		return nil
	}
	return s.engine.WriteFileToContainer(ctx, containerID, filePath, content)
}

func (s *Service) CreateInstance(userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.CreateInstanceWithContext(context.Background(), userID, challengeID)
}

func (s *Service) CreateInstanceWithContext(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 检查用户并发实例数
	instances, err := s.repo.FindByUserIDWithContext(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(instances) >= s.config.MaxConcurrentPerUser {
		return nil, errcode.ErrInstanceLimitExceeded
	}

	// 创建实例记录
	instance := &model.Instance{
		UserID:      userID,
		ChallengeID: challengeID,
		ContainerID: fmt.Sprintf("container-%d-%d", userID, time.Now().Unix()),
		Status:      model.InstanceStatusCreating,
		ExpiresAt:   time.Now().Add(s.config.DefaultTTL),
		MaxExtends:  s.config.MaxExtends,
	}

	if err := s.repo.CreateWithContext(ctx, instance); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	instance.Status = model.InstanceStatusRunning
	instance.AccessURL = fmt.Sprintf("http://localhost:3%04d", 1000+instance.ID)
	if err := s.repo.UpdateStatusWithContext(ctx, instance.ID, model.InstanceStatusRunning); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("创建实例",
		zap.Int64("user_id", userID),
		zap.Int64("challenge_id", challengeID),
		zap.Int64("instance_id", instance.ID),
		zap.Time("expires_at", instance.ExpiresAt))

	return toInstanceResp(instance), nil
}

func (s *Service) DestroyInstance(instanceID, userID int64) error {
	return s.DestroyInstanceWithContext(context.Background(), instanceID, userID)
}

func (s *Service) DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return errcode.ErrForbidden
	}

	s.logger.Info("销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("user_id", userID))

	return s.destroyManagedInstanceWithContext(ctx, instance)
}

func (s *Service) ExtendInstance(instanceID, userID int64) (*dto.InstanceResp, error) {
	return s.ExtendInstanceWithContext(context.Background(), instanceID, userID)
}

func (s *Service) ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return nil, errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning {
		return nil, errcode.ErrInstanceExpired
	}

	// 使用原子更新避免并发竞争
	if err := s.repo.AtomicExtendByIDWithContext(ctx, instanceID, s.config.MaxExtends, s.config.ExtendDuration); err != nil {
		return nil, err
	}

	updatedInstance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if updatedInstance == nil {
		return nil, errcode.ErrForbidden
	}

	s.logger.Info("延时实例",
		zap.Int64("instance_id", instanceID),
		zap.Int("extend_count", instance.ExtendCount+1),
		zap.Time("new_expires_at", instance.ExpiresAt.Add(s.config.ExtendDuration)))

	return toInstanceResp(updatedInstance), nil
}

func (s *Service) GetUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
	return s.GetUserInstancesWithContext(context.Background(), userID)
}

func (s *Service) GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	instances, err := s.repo.ListVisibleByUser(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.InstanceInfo, len(instances))
	for i, inst := range instances {
		result[i] = toInstanceInfo(inst, time.Now())
	}
	return result, nil
}

func (s *Service) GetAccessURL(instanceID, userID int64) (string, error) {
	return s.GetAccessURLWithContext(context.Background(), instanceID, userID)
}

func (s *Service) GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return "", errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return "", errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning || strings.TrimSpace(instance.AccessURL) == "" {
		return "", errcode.ErrInstanceExpired
	}
	return instance.AccessURL, nil
}

func (s *Service) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	filter := TeacherInstanceFilter{}
	if query != nil {
		filter.ClassName = strings.TrimSpace(query.ClassName)
		filter.Keyword = strings.TrimSpace(query.Keyword)
		filter.StudentNo = strings.TrimSpace(query.StudentNo)
	}

	if requesterRole != model.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return nil, errcode.ErrUnauthorized
		}

		className := strings.TrimSpace(requester.ClassName)
		if className == "" {
			return []dto.TeacherInstanceItem{}, nil
		}
		if filter.ClassName != "" && filter.ClassName != className {
			return nil, errcode.ErrForbidden
		}
		filter.ClassName = className
	}

	items, err := s.repo.ListTeacherInstances(ctx, filter)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	for idx := range items {
		items[idx].RemainingTime = calculateRemainingTime(items[idx].ExpiresAt, now)
	}

	return items, nil
}

func (s *Service) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	instance, err := s.repo.FindByID(instanceID)
	if err != nil {
		return errcode.ErrInstanceNotFound
	}

	if requesterRole != model.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return errcode.ErrUnauthorized
		}

		owner, err := s.repo.FindUserByID(ctx, instance.UserID)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if owner == nil || strings.TrimSpace(owner.ClassName) == "" || owner.ClassName != requester.ClassName {
			return errcode.ErrForbidden
		}
	}

	s.logger.Info("教师销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("requester_id", requesterID),
		zap.String("requester_role", requesterRole))

	return s.destroyManagedInstanceWithContext(ctx, instance)
}

func (s *Service) CleanExpiredInstances(ctx context.Context) error {
	instances, err := s.repo.FindExpired()
	if err != nil {
		return err
	}

	for _, inst := range instances {
		s.logger.Info("清理过期实例", zap.Int64("instance_id", inst.ID))

		if err := s.CleanupRuntimeWithContext(ctx, inst); err != nil {
			s.logger.Warn("清理过期实例运行时失败", zap.Int64("instance_id", inst.ID), zap.Error(err))
			continue
		}
		if err := s.repo.UpdateStatusAndReleasePort(inst.ID, model.InstanceStatusExpired); err != nil {
			s.logger.Warn("更新过期实例状态并释放端口失败", zap.Int64("instance_id", inst.ID), zap.Int("host_port", inst.HostPort), zap.Error(err))
		}
	}
	return nil
}

func (s *Service) CleanupOrphans(ctx context.Context) error {
	if s.engine == nil {
		s.logger.Debug("跳过孤儿容器清理，Docker 引擎未启用")
		return nil
	}

	managedContainers, err := s.engine.ListManagedContainers(ctx, managedByFilter())
	if err != nil {
		return err
	}
	activeContainerIDs, err := s.repo.ListActiveContainerIDs()
	if err != nil {
		return err
	}

	activeSet := make(map[string]struct{}, len(activeContainerIDs))
	for _, containerID := range activeContainerIDs {
		activeSet[containerID] = struct{}{}
	}

	orphanContainers := selectOrphanContainers(managedContainers, activeSet, s.config.OrphanGracePeriod, time.Now())
	for _, orphan := range orphanContainers {
		if err := s.RemoveContainerWithContext(ctx, orphan.ID); err != nil {
			s.logger.Warn("删除孤儿容器失败",
				zap.String("container_id", orphan.ID),
				zap.String("container_name", orphan.Name),
				zap.Error(err))
			continue
		}
		s.logger.Warn("已清理孤儿容器",
			zap.String("container_id", orphan.ID),
			zap.String("container_name", orphan.Name),
			zap.Time("created_at", orphan.CreatedAt))
	}

	return nil
}

func (s *Service) allocatePort() (int, error) {
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

func managedContainerLabels() map[string]string {
	return map[string]string{
		managedByLabelKey:         managedByLabelValue,
		challengeInstanceLabelKey: challengeInstanceLabelValue,
	}
}

func managedNetworkLabels() map[string]string {
	return map[string]string{
		managedByLabelKey:         managedByLabelValue,
		challengeInstanceLabelKey: challengeInstanceLabelValue,
	}
}

func buildManagedNetworkName(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		trimmed = model.TopologyDefaultNetworkKey
	}
	return fmt.Sprintf("%s%s-%d", managedNetworkNamePrefix, trimmed, time.Now().UnixNano())
}

func managedByFilter() string {
	return fmt.Sprintf("%s=%s", managedByLabelKey, managedByLabelValue)
}

func selectOrphanContainers(
	managedContainers []ManagedContainer,
	activeContainerIDs map[string]struct{},
	gracePeriod time.Duration,
	now time.Time,
) []ManagedContainer {
	orphanContainers := make([]ManagedContainer, 0, len(managedContainers))
	for _, container := range managedContainers {
		if _, exists := activeContainerIDs[container.ID]; exists {
			continue
		}
		if !container.CreatedAt.IsZero() && now.Sub(container.CreatedAt) < gracePeriod {
			continue
		}
		orphanContainers = append(orphanContainers, container)
	}
	return orphanContainers
}

func (s *Service) cleanupTopologyResources(containerIDs []string, networkIDs []string) {
	for idx := len(containerIDs) - 1; idx >= 0; idx-- {
		_ = s.RemoveContainer(containerIDs[idx])
	}
	for idx := len(networkIDs) - 1; idx >= 0; idx-- {
		_ = s.RemoveNetwork(networkIDs[idx])
	}
}

func toInstanceResp(inst *model.Instance) *dto.InstanceResp {
	return &dto.InstanceResp{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		Status:           inst.Status,
		AccessURL:        inst.AccessURL,
		ExpiresAt:        inst.ExpiresAt,
		ExtendCount:      inst.ExtendCount,
		MaxExtends:       inst.MaxExtends,
		RemainingExtends: remainingExtends(inst),
		CreatedAt:        inst.CreatedAt,
	}
}

func toInstanceInfo(inst UserVisibleInstanceRow, now time.Time) *dto.InstanceInfo {
	return &dto.InstanceInfo{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		ChallengeTitle:   inst.ChallengeTitle,
		Category:         inst.Category,
		Difficulty:       inst.Difficulty,
		FlagType:         inst.FlagType,
		Status:           inst.Status,
		AccessURL:        inst.AccessURL,
		ExpiresAt:        inst.ExpiresAt,
		RemainingTime:    calculateRemainingTime(inst.ExpiresAt, now),
		ExtendCount:      inst.ExtendCount,
		MaxExtends:       inst.MaxExtends,
		RemainingExtends: remainingExtends(&model.Instance{MaxExtends: inst.MaxExtends, ExtendCount: inst.ExtendCount}),
		CreatedAt:        inst.CreatedAt,
	}
}

func remainingExtends(inst *model.Instance) int {
	remaining := inst.MaxExtends - inst.ExtendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *Service) destroyManagedInstance(instance *model.Instance) error {
	return s.destroyManagedInstanceWithContext(context.Background(), instance)
}

func (s *Service) destroyManagedInstanceWithContext(ctx context.Context, instance *model.Instance) error {
	if err := s.CleanupRuntimeWithContext(ctx, instance); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.UpdateStatusAndReleasePort(instance.ID, model.InstanceStatusStopped); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func managedContainerIDs(instance *model.Instance) []string {
	if instance == nil {
		return nil
	}
	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil || len(details.Containers) == 0 {
		if instance.ContainerID == "" {
			return nil
		}
		return []string{instance.ContainerID}
	}
	result := make([]string, 0, len(details.Containers))
	seen := make(map[string]struct{}, len(details.Containers))
	for _, item := range details.Containers {
		if item.ContainerID == "" {
			continue
		}
		if _, exists := seen[item.ContainerID]; exists {
			continue
		}
		seen[item.ContainerID] = struct{}{}
		result = append(result, item.ContainerID)
	}
	if len(result) == 0 && instance.ContainerID != "" {
		return []string{instance.ContainerID}
	}
	return result
}

func managedNetworkIDs(instance *model.Instance) []string {
	if instance == nil {
		return nil
	}
	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err == nil && len(details.Networks) > 0 {
		result := make([]string, 0, len(details.Networks))
		seen := make(map[string]struct{}, len(details.Networks))
		for _, item := range details.Networks {
			if item.NetworkID == "" {
				continue
			}
			if _, exists := seen[item.NetworkID]; exists {
				continue
			}
			seen[item.NetworkID] = struct{}{}
			result = append(result, item.NetworkID)
		}
		if len(result) > 0 {
			return result
		}
	}
	if instance.NetworkID == "" {
		return nil
	}
	return []string{instance.NetworkID}
}

func managedACLRules(instance *model.Instance) []model.InstanceRuntimeACLRule {
	if instance == nil {
		return nil
	}
	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil {
		return nil
	}
	if len(details.ACLRules) == 0 {
		return nil
	}
	return append([]model.InstanceRuntimeACLRule(nil), details.ACLRules...)
}

func normalizedCreateNetworks(networks []TopologyCreateNetwork) []TopologyCreateNetwork {
	if len(networks) == 0 {
		return []TopologyCreateNetwork{{Key: model.TopologyDefaultNetworkKey}}
	}
	return networks
}

func normalizedNodeNetworkKeys(keys []string, networks []TopologyCreateNetwork) []string {
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

func calculateRemainingTime(expiresAt, now time.Time) int64 {
	remaining := int64(expiresAt.Sub(now).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}
