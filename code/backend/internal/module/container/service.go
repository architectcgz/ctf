package container

import (
	"context"
	"fmt"
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

const (
	managedByLabelKey           = "managed-by"
	managedByLabelValue         = "ctf-platform"
	challengeInstanceLabelKey   = "ctf-component"
	challengeInstanceLabelValue = "challenge-instance"
	managedContainerNamePrefix  = "ctf-instance-"
	managedNetworkNamePrefix    = "ctf-net-"
)

type runtimeEngine interface {
	CreateNetwork(ctx context.Context, name string, labels map[string]string) (string, error)
	CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	ListManagedContainers(ctx context.Context, managedBy string) ([]ManagedContainer, error)
}

func NewService(repo *Repository, engine runtimeEngine, cfg *config.ContainerConfig, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		repo:   repo,
		engine: engine,
		config: cfg,
		logger: logger,
	}
}

func (s *Service) CreateContainer(ctx context.Context, imageName string, env map[string]string) (containerID, networkID string, port int, err error) {
	port, err = s.allocatePort()
	if err != nil {
		return "", "", 0, err
	}

	if s.engine == nil {
		select {
		case <-ctx.Done():
			return "", "", 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}

		containerID = fmt.Sprintf("ctf-%d", time.Now().UnixNano())
		return containerID, "", port, nil
	}

	networkName := buildManagedNetworkName()
	networkID, err = s.engine.CreateNetwork(ctx, networkName, managedNetworkLabels())
	if err != nil {
		return "", "", 0, err
	}

	containerID, err = s.engine.CreateContainer(ctx, &model.ContainerConfig{
		Image: imageName,
		Name:  buildManagedContainerName(),
		Env:   envMapToList(env),
		Ports: map[string]string{
			strconv.Itoa(s.config.DefaultExposedPort): strconv.Itoa(port),
		},
		Labels:  managedContainerLabels(),
		Network: networkName,
	})
	if err != nil {
		_ = s.engine.RemoveNetwork(context.Background(), networkID)
		return "", "", 0, err
	}
	if err := s.engine.StartContainer(ctx, containerID); err != nil {
		_ = s.engine.RemoveContainer(context.Background(), containerID, true)
		_ = s.engine.RemoveNetwork(context.Background(), networkID)
		return "", "", 0, err
	}

	return containerID, networkID, port, nil
}

// RemoveContainer 删除容器
func (s *Service) RemoveContainer(containerID string) error {
	if s.engine == nil {
		s.logger.Info("删除容器（降级模拟）", zap.String("container_id", containerID))
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = s.engine.StopContainer(ctx, containerID, 5*time.Second)
	if err := s.engine.RemoveContainer(ctx, containerID, true); err != nil {
		return err
	}

	s.logger.Info("删除容器", zap.String("container_id", containerID))
	return nil
}

// RemoveNetwork 删除网络
func (s *Service) RemoveNetwork(networkID string) error {
	if networkID == "" {
		return nil
	}
	if s.engine == nil {
		s.logger.Info("删除网络（降级跳过）", zap.String("network_id", networkID))
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.engine.RemoveNetwork(ctx, networkID); err != nil {
		return err
	}

	s.logger.Info("删除网络", zap.String("network_id", networkID))
	return nil
}

func (s *Service) CreateInstance(userID, challengeID int64) (*dto.InstanceResp, error) {
	// 检查用户并发实例数
	instances, err := s.repo.FindByUserID(userID)
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

	if err := s.repo.Create(instance); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	instance.Status = model.InstanceStatusRunning
	instance.AccessURL = fmt.Sprintf("http://localhost:3%04d", 1000+instance.ID)
	if err := s.repo.UpdateStatus(instance.ID, model.InstanceStatusRunning); err != nil {
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
	instance, err := s.repo.FindByID(instanceID)
	if err != nil {
		return errcode.ErrInstanceNotFound
	}
	if instance.UserID != userID {
		return errcode.ErrForbidden
	}

	s.logger.Info("销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("user_id", userID))

	return s.destroyManagedInstance(instance)
}

func (s *Service) ExtendInstance(instanceID, userID int64) error {
	instance, err := s.repo.FindByID(instanceID)
	if err != nil {
		return errcode.ErrInstanceNotFound
	}
	if instance.UserID != userID {
		return errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning {
		return errcode.ErrInstanceExpired
	}

	// 使用原子更新避免并发竞争
	if err := s.repo.AtomicExtend(instanceID, userID, s.config.MaxExtends, s.config.ExtendDuration); err != nil {
		return err
	}

	s.logger.Info("延时实例",
		zap.Int64("instance_id", instanceID),
		zap.Int("extend_count", instance.ExtendCount+1),
		zap.Time("new_expires_at", instance.ExpiresAt.Add(s.config.ExtendDuration)))

	return nil
}

func (s *Service) GetUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
	instances, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.InstanceInfo, len(instances))
	for i, inst := range instances {
		result[i] = toInstanceInfo(inst)
	}
	return result, nil
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

	return s.destroyManagedInstance(instance)
}

func (s *Service) CleanExpiredInstances(ctx context.Context) error {
	instances, err := s.repo.FindExpired()
	if err != nil {
		return err
	}

	for _, inst := range instances {
		s.logger.Info("清理过期实例", zap.Int64("instance_id", inst.ID))

		if inst.ContainerID != "" {
			if err := s.RemoveContainer(inst.ContainerID); err != nil {
				s.logger.Warn("删除过期容器失败", zap.Int64("instance_id", inst.ID), zap.Error(err))
			}
		}
		if inst.NetworkID != "" {
			if err := s.RemoveNetwork(inst.NetworkID); err != nil {
				s.logger.Warn("删除过期网络失败", zap.Int64("instance_id", inst.ID), zap.Error(err))
			}
		}

		s.repo.UpdateStatus(inst.ID, model.InstanceStatusExpired)
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
		if err := s.RemoveContainer(orphan.ID); err != nil {
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

func buildManagedNetworkName() string {
	return fmt.Sprintf("%s%d", managedNetworkNamePrefix, time.Now().UnixNano())
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

func toInstanceResp(inst *model.Instance) *dto.InstanceResp {
	return &dto.InstanceResp{
		ID:          inst.ID,
		ChallengeID: inst.ChallengeID,
		Status:      inst.Status,
		AccessURL:   inst.AccessURL,
		ExpiresAt:   inst.ExpiresAt,
		ExtendCount: inst.ExtendCount,
		MaxExtends:  inst.MaxExtends,
		CreatedAt:   inst.CreatedAt,
	}
}

func toInstanceInfo(inst *model.Instance) *dto.InstanceInfo {
	return &dto.InstanceInfo{
		ID:            inst.ID,
		ChallengeID:   inst.ChallengeID,
		Status:        inst.Status,
		AccessURL:     inst.AccessURL,
		ExpiresAt:     inst.ExpiresAt,
		RemainingTime: calculateRemainingTime(inst.ExpiresAt, time.Now()),
		ExtendCount:   inst.ExtendCount,
		MaxExtends:    inst.MaxExtends,
		CreatedAt:     inst.CreatedAt,
	}
}

func (s *Service) destroyManagedInstance(instance *model.Instance) error {
	if instance.ContainerID != "" {
		if err := s.RemoveContainer(instance.ContainerID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	if instance.NetworkID != "" {
		if err := s.RemoveNetwork(instance.NetworkID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	if err := s.repo.UpdateStatus(instance.ID, model.InstanceStatusStopped); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func calculateRemainingTime(expiresAt, now time.Time) int64 {
	remaining := int64(expiresAt.Sub(now).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}
