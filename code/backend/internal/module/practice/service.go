package practice

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

const errMsgChallengeNoTarget = "该题目不需要靶机实例"

type AssessmentService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

type ScoreUpdater interface {
	UpdateUserScoreWithContext(ctx context.Context, userID int64) error
	lockTimeout() time.Duration
}

type imageStore interface {
	FindByID(id int64) (*model.Image, error)
}

type runtimeInstanceService interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type instanceRepository interface {
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

type Service struct {
	repo              *Repository
	challengeRepo     challenge.PracticeChallengeContract
	imageRepo         imageStore
	instanceRepo      instanceRepository
	runtimeService    runtimeInstanceService
	scoreService      ScoreUpdater
	assessmentService AssessmentService
	redis             *redis.Client
	config            *config.Config
	logger            *zap.Logger
	eventBus          platformevents.Bus
	baseCtx           context.Context
	cancel            context.CancelFunc
	tasks             sync.WaitGroup
}

func (s *Service) SetEventBus(bus platformevents.Bus) *Service {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func NewService(
	repo *Repository,
	challengeRepo challenge.PracticeChallengeContract,
	imageRepo imageStore,
	instanceRepo instanceRepository,
	runtimeService runtimeInstanceService,
	scoreService ScoreUpdater,
	assessmentService AssessmentService,
	redis *redis.Client,
	config *config.Config,
	logger *zap.Logger,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	baseCtx, cancel := context.WithCancel(context.Background())
	return &Service{
		repo:              repo,
		challengeRepo:     challengeRepo,
		imageRepo:         imageRepo,
		instanceRepo:      instanceRepo,
		runtimeService:    runtimeService,
		scoreService:      scoreService,
		assessmentService: assessmentService,
		redis:             redis,
		config:            config,
		logger:            logger,
		baseCtx:           baseCtx,
		cancel:            cancel,
	}
}

func (s *Service) StartChallenge(userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.StartChallengeWithContext(context.Background(), userID, challengeID)
}

func (s *Service) StartChallengeWithContext(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.startPersonalChallenge(ctx, userID, challengeID)
}

func (s *Service) StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*dto.InstanceResp, error) {
	scope, err := s.resolveContestInstanceScope(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, err
	}
	return s.startChallengeWithScope(ctx, userID, challengeID, scope)
}

func (s *Service) startPersonalChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.startChallengeWithScope(ctx, userID, challengeID, instanceScope{
		flagSubjectID: userID,
	})
}

type instanceScope struct {
	contestID     *int64
	teamID        *int64
	flagSubjectID int64
}

func (s *Service) startChallengeWithScope(ctx context.Context, userID, challengeID int64, scope instanceScope) (*dto.InstanceResp, error) {
	chal, err := s.challengeRepo.FindByID(challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if chal.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}
	if chal.ImageID == 0 {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New(errMsgChallengeNoTarget))
	}

	flag, nonce, err := s.buildInstanceFlag(scope.flagSubjectID, challengeID, chal)
	if err != nil {
		return nil, err
	}

	var (
		instance *model.Instance
		reused   bool
	)
	if err := s.repo.WithinTransaction(ctx, func(txRepo *Repository) error {
		if err := txRepo.LockInstanceScope(userID, scope); err != nil {
			return err
		}

		existingInstance, err := txRepo.FindScopedExistingInstance(userID, challengeID, scope)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if existingInstance != nil {
			instance = existingInstance
			reused = true
			return nil
		}

		runningCount, err := txRepo.CountScopedRunningInstances(userID, scope)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if runningCount >= s.config.Container.MaxConcurrentPerUser {
			s.logger.Warn("实例数量超限",
				zap.Int64("user_id", userID),
				zap.Int64("challenge_id", challengeID),
				zap.Int("current", runningCount),
				zap.Int("limit", s.config.Container.MaxConcurrentPerUser))
			return errcode.ErrInstanceLimitExceeded
		}

		hostPort, err := txRepo.ReserveAvailablePort(s.config.Container.PortRangeStart, s.config.Container.PortRangeEnd)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}

		instance = &model.Instance{
			UserID:      userID,
			ContestID:   scope.contestID,
			TeamID:      scope.teamID,
			ChallengeID: challengeID,
			HostPort:    hostPort,
			Status:      model.InstanceStatusCreating,
			Nonce:       nonce,
			ExpiresAt:   time.Now().Add(s.config.Container.DefaultTTL),
			MaxExtends:  s.config.Container.MaxExtends,
		}
		if err := txRepo.CreateInstance(instance); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if err := txRepo.BindReservedPort(hostPort, instance.ID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if reused {
		return toInstanceResp(instance), nil
	}

	createCtx, cancel := context.WithTimeout(ctx, s.config.Container.CreateTimeout)
	defer cancel()

	if err := s.createContainer(createCtx, instance, chal, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		s.markInstanceFailed(instance)
		return nil, err
	}

	instance.Status = model.InstanceStatusRunning
	if err := s.instanceRepo.UpdateRuntime(instance); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err))
		s.markInstanceFailed(instance)
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("实例启动成功",
		zap.Int64("user_id", userID),
		zap.Int64("challenge_id", challengeID),
		zap.Int64("instance_id", instance.ID))

	return toInstanceResp(instance), nil
}

func (s *Service) markInstanceFailed(instance *model.Instance) {
	if instance == nil {
		return
	}
	if err := s.runtimeService.CleanupRuntime(instance); err != nil {
		s.logger.Warn("清理失败实例运行时资源失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	if err := s.instanceRepo.UpdateStatusAndReleasePort(instance.ID, model.InstanceStatusFailed); err != nil {
		s.logger.Warn("更新失败实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
	}
}

func (s *Service) SubmitFlag(userID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	return s.SubmitFlagWithContext(context.Background(), userID, challengeID, flag)
}

func (s *Service) SubmitFlagWithContext(ctx context.Context, userID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	challengeItem, err := s.challengeRepo.FindByID(challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotFound
		}
		s.logger.Error("查询靶场失败", zap.Int64("challenge_id", challengeID), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	if _, err := s.repo.FindCorrectSubmission(userID, challengeID); err == nil {
		return nil, errcode.ErrAlreadySolved
	}

	rateLimitKey := fmt.Sprintf("%s:submit:%d:%d", s.config.RateLimit.RedisKeyPrefix, userID, challengeID)
	count, err := s.redis.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		s.logger.Error("提交限流失败", zap.String("key", rateLimitKey), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if count == 1 {
		_ = s.redis.Expire(ctx, rateLimitKey, s.config.RateLimit.FlagSubmit.Window).Err()
	}
	if count > int64(s.config.RateLimit.FlagSubmit.Limit) {
		return nil, errcode.ErrSubmitTooFrequent
	}

	isCorrect, err := s.validateSubmittedFlag(userID, challengeItem, flag)
	if err != nil {
		return nil, err
	}

	submission := &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		Flag:        "",
		IsCorrect:   isCorrect,
		SubmittedAt: time.Now(),
	}
	if err := s.repo.CreateSubmission(submission); err != nil {
		if isCorrect && s.repo.IsUniqueViolation(err) {
			return nil, errcode.ErrAlreadySolved
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if isCorrect {
		cacheKey := constants.UserProgressKey(userID)
		if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
			s.logger.Warn("删除进度缓存失败", zap.Int64("user_id", userID), zap.Error(err))
		}
		s.publishWeakEvent(ctx, platformevents.Event{
			Name: EventFlagAccepted,
			Payload: FlagAcceptedEvent{
				UserID:      userID,
				ChallengeID: challengeID,
				Dimension:   challengeItem.Category,
				Points:      challengeItem.Points,
				OccurredAt:  submission.SubmittedAt,
			},
		})
	}

	resp := &dto.SubmissionResp{
		IsCorrect:   isCorrect,
		SubmittedAt: submission.SubmittedAt,
	}
	if isCorrect {
		resp.Message = "恭喜你，Flag 正确！"
		resp.Points = challengeItem.Points
		if s.scoreService != nil {
			s.triggerScoreUpdate(userID)
		}
	} else {
		resp.Message = "Flag 错误，请重试"
	}

	return resp, nil
}

func (s *Service) UnlockHint(userID, challengeID int64, level int) (*dto.UnlockHintResp, error) {
	challengeItem, err := s.challengeRepo.FindByID(challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	hint, err := s.challengeRepo.FindHintByLevel(challengeID, level)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if err := s.challengeRepo.CreateHintUnlock(&model.ChallengeHintUnlock{
		UserID:          userID,
		ChallengeID:     challengeID,
		ChallengeHintID: hint.ID,
		UnlockedAt:      time.Now(),
	}); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.publishWeakEvent(context.Background(), platformevents.Event{
		Name: EventHintUnlocked,
		Payload: HintUnlockedEvent{
			UserID:      userID,
			ChallengeID: challengeID,
			Dimension:   challengeItem.Category,
			HintLevel:   hint.Level,
			OccurredAt:  time.Now(),
		},
	})

	return &dto.UnlockHintResp{
		Hint: &dto.ChallengeHintResp{
			ID:         hint.ID,
			Level:      hint.Level,
			Title:      hint.Title,
			CostPoints: hint.CostPoints,
			IsUnlocked: true,
			Content:    hint.Content,
		},
	}, nil
}

func (s *Service) resolveContestInstanceScope(ctx context.Context, userID, contestID, challengeID int64) (instanceScope, error) {
	if s.repo == nil {
		return instanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice repository is nil"))
	}

	contest, err := s.repo.FindContestByIDWithContext(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instanceScope{}, errcode.ErrContestNotFound
		}
		return instanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch contest.Status {
	case model.ContestStatusRunning, model.ContestStatusFrozen:
	default:
		if contest.Status == model.ContestStatusEnded {
			return instanceScope{}, errcode.ErrContestEnded
		}
		return instanceScope{}, errcode.ErrContestNotRunning
	}

	contestChallenge, err := s.repo.FindContestChallengeWithContext(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instanceScope{}, errcode.ErrChallengeNotInContest
		}
		return instanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if !contestChallenge.IsVisible {
		return instanceScope{}, errcode.ErrContestChallengeVisible
	}

	registration, err := s.repo.FindContestRegistrationWithContext(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instanceScope{}, errcode.ErrNotRegistered
		}
		return instanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch registration.Status {
	case model.ContestRegistrationStatusApproved:
	case model.ContestRegistrationStatusPending:
		return instanceScope{}, errcode.ErrContestRegistrationPending
	default:
		return instanceScope{}, errcode.ErrRegistrationNotApproved
	}

	contestIDCopy := contestID
	scope := instanceScope{
		contestID:     &contestIDCopy,
		flagSubjectID: userID,
	}
	if contest.Mode == model.ContestModeAWD {
		if registration.TeamID == nil || *registration.TeamID <= 0 {
			return instanceScope{}, errcode.ErrAWDTeamRequired
		}
		teamID := *registration.TeamID
		scope.teamID = &teamID
		scope.flagSubjectID = teamID
		return scope, nil
	}

	return scope, nil
}

func (s *Service) buildInstanceFlag(subjectID, challengeID int64, chal *model.Challenge) (string, string, error) {
	switch chal.FlagType {
	case model.FlagTypeDynamic:
		nonce, err := crypto.GenerateNonce()
		if err != nil {
			return "", "", errcode.ErrInternal.WithCause(err)
		}
		if s.config.Container.FlagGlobalSecret == "" {
			return "", "", errcode.ErrInternal.WithCause(fmt.Errorf("flag global secret is empty"))
		}
		flag := crypto.GenerateDynamicFlag(subjectID, challengeID, s.config.Container.FlagGlobalSecret, nonce, chal.FlagPrefix)
		return flag, nonce, nil
	case model.FlagTypeStatic:
		return chal.FlagHash, "", nil
	default:
		return "", "", nil
	}
}

func (s *Service) validateSubmittedFlag(userID int64, challengeItem *model.Challenge, flag string) (bool, error) {
	if challengeItem.FlagType == model.FlagTypeStatic {
		inputHash := crypto.HashStaticFlag(flag, challengeItem.FlagSalt)
		return crypto.ValidateFlag(inputHash, challengeItem.FlagHash), nil
	}

	instance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeItem.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil || instance.Nonce == "" || s.config.Container.FlagGlobalSecret == "" {
		return false, nil
	}

	expectedFlag := crypto.GenerateDynamicFlag(userID, challengeItem.ID, s.config.Container.FlagGlobalSecret, instance.Nonce, challengeItem.FlagPrefix)
	return crypto.ValidateFlag(flag, expectedFlag), nil
}

func (s *Service) createContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, flag string) error {
	topology, err := s.challengeRepo.FindChallengeTopologyByChallengeID(chal.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	if topology == nil {
		return s.createSingleContainer(ctx, instance, chal, flag)
	}

	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	request, err := s.buildTopologyCreateRequest(instance.HostPort, chal, topology.EntryNodeKey, spec, flag)
	if err != nil {
		return err
	}
	result, err := s.runtimeService.CreateTopology(ctx, request)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(result.RuntimeDetails)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	instance.ContainerID = result.PrimaryContainerID
	instance.NetworkID = result.NetworkID
	instance.RuntimeDetails = runtimeDetails
	instance.AccessURL = result.AccessURL
	return nil
}

func (s *Service) createSingleContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, flag string) error {
	imageItem, err := s.imageRepo.FindByID(chal.ImageID)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	if imageItem.Status != model.ImageStatusAvailable {
		return errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("image %d is not available", imageItem.ID))
	}

	env := map[string]string{
		"FLAG": flag,
	}

	imageRef := fmt.Sprintf("%s:%s", imageItem.Name, imageItem.Tag)
	containerID, networkID, hostPort, servicePort, err := s.runtimeService.CreateContainer(ctx, imageRef, env, instance.HostPort)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Containers: []model.InstanceRuntimeContainer{
			{
				NodeKey:      "default",
				ContainerID:  containerID,
				ServicePort:  servicePort,
				HostPort:     hostPort,
				IsEntryPoint: true,
			},
		},
	})
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	instance.ContainerID = containerID
	instance.NetworkID = networkID
	instance.RuntimeDetails = runtimeDetails
	instance.AccessURL = fmt.Sprintf("http://%s:%d", s.config.Container.PublicHost, hostPort)
	return nil
}

func (s *Service) buildTopologyCreateRequest(
	reservedHostPort int,
	chal *model.Challenge,
	entryNodeKey string,
	spec model.TopologySpec,
	flag string,
) (*TopologyCreateRequest, error) {
	if len(spec.Nodes) == 0 {
		return nil, errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("challenge topology has no nodes"))
	}

	defaultImageRef, err := s.resolveAvailableImageRef(chal.ImageID)
	if err != nil {
		return nil, err
	}

	request := &TopologyCreateRequest{
		ReservedHostPort: reservedHostPort,
		Networks:         make([]TopologyCreateNetwork, 0),
		Nodes:            make([]TopologyCreateNode, 0, len(spec.Nodes)),
		Policies:         append([]model.TopologyTrafficPolicy(nil), spec.Policies...),
	}
	runtimePlan := buildRuntimeTopologyPlan(spec)
	request.Networks = append(request.Networks, runtimePlan.Networks...)
	for _, node := range spec.Nodes {
		imageRef := defaultImageRef
		if node.ImageID > 0 {
			imageRef, err = s.resolveAvailableImageRef(node.ImageID)
			if err != nil {
				return nil, err
			}
		}

		env := make(map[string]string, len(node.Env)+1)
		for key, value := range node.Env {
			env[key] = value
		}
		if node.InjectFlag {
			env["FLAG"] = flag
		}

		var resources *model.ResourceLimits
		if node.Resources != nil {
			resources = &model.ResourceLimits{
				CPUQuota:  node.Resources.CPUQuota,
				Memory:    node.Resources.MemoryMB * 1024 * 1024,
				PidsLimit: node.Resources.PidsLimit,
			}
		}

		request.Nodes = append(request.Nodes, TopologyCreateNode{
			Key:          node.Key,
			Image:        imageRef,
			Env:          env,
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.Key == entryNodeKey,
			NetworkKeys:  append([]string(nil), runtimePlan.NodeNetworkKeys[node.Key]...),
			Resources:    resources,
		})
	}

	return request, nil
}

func (s *Service) resolveAvailableImageRef(imageID int64) (string, error) {
	imageItem, err := s.imageRepo.FindByID(imageID)
	if err != nil {
		return "", errcode.ErrContainerCreateFailed.WithCause(err)
	}
	if imageItem.Status != model.ImageStatusAvailable {
		return "", errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("image %d is not available", imageItem.ID))
	}
	return fmt.Sprintf("%s:%s", imageItem.Name, imageItem.Tag), nil
}

func (s *Service) triggerAssessmentUpdate(userID int64, dimension string) {
	if s.assessmentService == nil || !model.IsValidDimension(dimension) {
		return
	}

	s.runAsyncTask(func(ctx context.Context) {
		timer := time.NewTimer(s.config.Assessment.IncrementalUpdateDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		updateCtx, cancel := context.WithTimeout(ctx, s.config.Assessment.IncrementalUpdateTimeout)
		defer cancel()

		if err := s.assessmentService.UpdateSkillProfileForDimension(updateCtx, userID, dimension); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("更新能力画像失败",
				zap.Int64("user_id", userID),
				zap.String("dimension", dimension),
				zap.Error(err))
		}
	})
}

func (s *Service) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})
	go func() {
		s.tasks.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) triggerScoreUpdate(userID int64) {
	if s.scoreService == nil {
		return
	}

	s.runAsyncTask(func(ctx context.Context) {
		scoreCtx := ctx
		cancel := func() {}
		if timeout := s.scoreService.lockTimeout(); timeout > 0 {
			scoreCtx, cancel = context.WithTimeout(ctx, timeout)
		}
		defer cancel()

		if err := s.scoreService.UpdateUserScoreWithContext(scoreCtx, userID); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("更新用户得分失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	})
}

func (s *Service) runAsyncTask(fn func(context.Context)) {
	if s.baseCtx == nil {
		return
	}

	s.tasks.Add(1)
	go func() {
		defer s.tasks.Done()

		select {
		case <-s.baseCtx.Done():
			return
		default:
		}

		fn(s.baseCtx)
	}()
}

func (s *Service) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
	if s.eventBus == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.logger.Warn("publish_practice_event_failed", zap.String("event", evt.Name), zap.Error(err))
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

func remainingExtends(inst *model.Instance) int {
	remaining := inst.MaxExtends - inst.ExtendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
