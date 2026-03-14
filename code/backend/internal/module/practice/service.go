package practice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
	"ctf-platform/internal/module/container"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type AssessmentService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

type Service struct {
	db                *gorm.DB
	repo              *Repository
	challengeRepo     *challenge.Repository
	imageRepo         *challenge.ImageRepository
	instanceRepo      *container.Repository
	containerService  *container.Service
	scoreService      *ScoreService
	assessmentService AssessmentService
	redis             *redis.Client
	config            *config.Config
	logger            *zap.Logger
}

func NewService(
	db *gorm.DB,
	repo *Repository,
	challengeRepo *challenge.Repository,
	imageRepo *challenge.ImageRepository,
	instanceRepo *container.Repository,
	containerService *container.Service,
	scoreService *ScoreService,
	assessmentService AssessmentService,
	redis *redis.Client,
	config *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		db:                db,
		repo:              repo,
		challengeRepo:     challengeRepo,
		imageRepo:         imageRepo,
		instanceRepo:      instanceRepo,
		containerService:  containerService,
		scoreService:      scoreService,
		assessmentService: assessmentService,
		redis:             redis,
		config:            config,
		logger:            logger,
	}
}

func (s *Service) StartChallenge(userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.startPersonalChallenge(context.Background(), userID, challengeID)
}

func (s *Service) StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*dto.InstanceResp, error) {
	scope, existingInstance, runningCount, err := s.resolveContestInstanceScope(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, err
	}
	if existingInstance != nil {
		return toInstanceResp(existingInstance), nil
	}
	if runningCount >= s.config.Container.MaxConcurrentPerUser {
		s.logger.Warn("竞赛实例数量超限",
			zap.Int64("user_id", userID),
			zap.Int64("contest_id", contestID),
			zap.Int("current", runningCount),
			zap.Int("limit", s.config.Container.MaxConcurrentPerUser))
		return nil, errcode.ErrInstanceLimitExceeded
	}
	return s.startChallengeWithScope(ctx, userID, challengeID, scope)
}

func (s *Service) startPersonalChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	existingInstance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
	if err == nil && existingInstance != nil {
		return toInstanceResp(existingInstance), nil
	}

	instances, err := s.instanceRepo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(instances) >= s.config.Container.MaxConcurrentPerUser {
		s.logger.Warn("用户实例数量超限",
			zap.Int64("user_id", userID),
			zap.Int("current", len(instances)),
			zap.Int("limit", s.config.Container.MaxConcurrentPerUser))
		return nil, errcode.ErrInstanceLimitExceeded
	}
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

	flag, nonce, err := s.buildInstanceFlag(scope.flagSubjectID, challengeID, chal)
	if err != nil {
		return nil, err
	}

	instance := &model.Instance{
		UserID:      userID,
		ContestID:   scope.contestID,
		TeamID:      scope.teamID,
		ChallengeID: challengeID,
		Status:      model.InstanceStatusCreating,
		Nonce:       nonce,
		ExpiresAt:   time.Now().Add(s.config.Container.DefaultTTL),
		MaxExtends:  s.config.Container.MaxExtends,
	}

	if err := s.instanceRepo.Create(instance); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	createCtx, cancel := context.WithTimeout(ctx, s.config.Container.CreateTimeout)
	defer cancel()

	if err := s.createContainer(createCtx, instance, chal, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		if instance.NetworkID != "" {
			_ = s.containerService.RemoveNetwork(instance.NetworkID)
		}
		if instance.ContainerID != "" {
			_ = s.containerService.RemoveContainer(instance.ContainerID)
		}
		_ = s.instanceRepo.UpdateStatus(instance.ID, model.InstanceStatusFailed)
		return nil, err
	}

	instance.Status = model.InstanceStatusRunning
	if err := s.instanceRepo.UpdateRuntime(instance); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("实例启动成功",
		zap.Int64("user_id", userID),
		zap.Int64("challenge_id", challengeID),
		zap.Int64("instance_id", instance.ID))

	return toInstanceResp(instance), nil
}

func (s *Service) SubmitFlag(userID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
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

	ctx := context.Background()
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
		if err := s.redis.Del(ctx, cacheKey, rediskeys.RecommendationKey(userID)).Err(); err != nil {
			s.logger.Warn("删除进度缓存失败", zap.Int64("user_id", userID), zap.Error(err))
		}
		s.triggerAssessmentUpdate(userID, challengeItem.Category)
	}

	resp := &dto.SubmissionResp{
		IsCorrect:   isCorrect,
		SubmittedAt: submission.SubmittedAt,
	}
	if isCorrect {
		resp.Message = "恭喜你，Flag 正确！"
		resp.Points = challengeItem.Points
		if s.scoreService != nil {
			go func() {
				if err := s.scoreService.UpdateUserScore(userID); err != nil {
					s.logger.Error("更新用户得分失败", zap.Int64("user_id", userID), zap.Error(err))
				}
			}()
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

func (s *Service) resolveContestInstanceScope(ctx context.Context, userID, contestID, challengeID int64) (instanceScope, *model.Instance, int, error) {
	if s.db == nil {
		return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(fmt.Errorf("contest instance db is nil"))
	}

	var contest model.Contest
	if err := s.db.WithContext(ctx).Where("id = ?", contestID).First(&contest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instanceScope{}, nil, 0, errcode.ErrContestNotFound
		}
		return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
	}
	switch contest.Status {
	case model.ContestStatusRunning, model.ContestStatusFrozen:
	default:
		if contest.Status == model.ContestStatusEnded {
			return instanceScope{}, nil, 0, errcode.ErrContestEnded
		}
		return instanceScope{}, nil, 0, errcode.ErrContestNotRunning
	}

	var contestChallenge model.ContestChallenge
	if err := s.db.WithContext(ctx).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&contestChallenge).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instanceScope{}, nil, 0, errcode.ErrChallengeNotInContest
		}
		return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
	}
	if !contestChallenge.IsVisible {
		return instanceScope{}, nil, 0, errcode.ErrContestChallengeVisible
	}

	var registration model.ContestRegistration
	if err := s.db.WithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return instanceScope{}, nil, 0, errcode.ErrNotRegistered
		}
		return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
	}
	switch registration.Status {
	case model.ContestRegistrationStatusApproved:
	case model.ContestRegistrationStatusPending:
		return instanceScope{}, nil, 0, errcode.ErrContestRegistrationPending
	default:
		return instanceScope{}, nil, 0, errcode.ErrRegistrationNotApproved
	}

	contestIDCopy := contestID
	scope := instanceScope{
		contestID:     &contestIDCopy,
		flagSubjectID: userID,
	}
	if contest.Mode == model.ContestModeAWD {
		if registration.TeamID == nil || *registration.TeamID <= 0 {
			return instanceScope{}, nil, 0, errcode.ErrAWDTeamRequired
		}
		teamID := *registration.TeamID
		scope.teamID = &teamID
		scope.flagSubjectID = teamID
		existing, err := s.instanceRepo.FindByContestTeamAndChallenge(contestID, teamID, challengeID)
		if err != nil {
			return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
		}
		instances, err := s.instanceRepo.FindByContestTeamID(contestID, teamID)
		if err != nil {
			return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
		}
		return scope, existing, len(instances), nil
	}

	existing, err := s.instanceRepo.FindByContestUserAndChallenge(contestID, userID, challengeID)
	if err != nil {
		return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
	}
	instances, err := s.instanceRepo.FindByContestUserID(contestID, userID)
	if err != nil {
		return instanceScope{}, nil, 0, errcode.ErrInternal.WithCause(err)
	}
	return scope, existing, len(instances), nil
}

func (s *Service) GetProgress(userID int64) (*dto.ProgressResp, error) {
	ctx := context.Background()
	cacheKey := constants.UserProgressKey(userID)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var resp dto.ProgressResp
		if json.Unmarshal([]byte(cached), &resp) == nil {
			return &resp, nil
		}
		s.logger.Warn("进度缓存反序列化失败", zap.Int64("user_id", userID))
	}

	totalScore, totalSolved, err := s.repo.GetUserProgress(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rank, err := s.repo.GetUserRank(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	categoryStats, err := s.repo.GetCategoryStats(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	difficultyStats, err := s.repo.GetDifficultyStats(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.ProgressResp{
		TotalScore:      totalScore,
		TotalSolved:     totalSolved,
		Rank:            rank,
		CategoryStats:   make([]dto.CategoryStat, len(categoryStats)),
		DifficultyStats: make([]dto.DifficultyStat, len(difficultyStats)),
	}
	for i, stat := range categoryStats {
		resp.CategoryStats[i] = dto.CategoryStat{
			Category: stat.Category,
			Solved:   stat.Solved,
			Total:    stat.Total,
		}
	}
	for i, stat := range difficultyStats {
		resp.DifficultyStats[i] = dto.DifficultyStat{
			Difficulty: stat.Difficulty,
			Solved:     stat.Solved,
			Total:      stat.Total,
		}
	}

	if data, err := json.Marshal(resp); err == nil {
		_ = s.redis.Set(ctx, cacheKey, data, s.config.Cache.ProgressTTL).Err()
	}

	return resp, nil
}

func (s *Service) GetTimeline(userID int64, limit, offset int) (*dto.TimelineResp, error) {
	events, err := s.repo.GetUserTimeline(userID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.TimelineResp{
		Events: make([]dto.TimelineEvent, len(events)),
	}
	for i, event := range events {
		resp.Events[i] = dto.TimelineEvent{
			Type:        event.Type,
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Timestamp:   event.Timestamp,
			IsCorrect:   event.IsCorrect,
			Points:      event.Points,
			Detail:      event.Detail,
		}
	}
	return resp, nil
}

func (s *Service) GetInstance(instanceID, userID int64) (*dto.InstanceInfo, error) {
	instance, err := s.instanceRepo.FindAccessibleByIDForUser(context.Background(), instanceID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return nil, errcode.ErrInstanceNotFound
	}

	return toInstanceInfo(instance), nil
}

func (s *Service) ListUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
	instances, err := s.instanceRepo.FindVisibleByUser(context.Background(), userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.InstanceInfo, len(instances))
	for i, inst := range instances {
		result[i] = toInstanceInfo(inst)
		if chal, err := s.challengeRepo.FindByID(inst.ChallengeID); err == nil {
			result[i].ChallengeName = chal.Title
		}
	}
	return result, nil
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

	request, err := s.buildTopologyCreateRequest(chal, topology.EntryNodeKey, spec, flag)
	if err != nil {
		return err
	}
	result, err := s.containerService.CreateTopology(ctx, request)
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
	containerID, networkID, hostPort, servicePort, err := s.containerService.CreateContainer(ctx, imageRef, env)
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
	chal *model.Challenge,
	entryNodeKey string,
	spec model.TopologySpec,
	flag string,
) (*container.TopologyCreateRequest, error) {
	if len(spec.Nodes) == 0 {
		return nil, errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("challenge topology has no nodes"))
	}

	defaultImageRef, err := s.resolveAvailableImageRef(chal.ImageID)
	if err != nil {
		return nil, err
	}

	request := &container.TopologyCreateRequest{
		Networks: make([]container.TopologyCreateNetwork, 0),
		Nodes:    make([]container.TopologyCreateNode, 0, len(spec.Nodes)),
		Policies: append([]model.TopologyTrafficPolicy(nil), spec.Policies...),
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

		request.Nodes = append(request.Nodes, container.TopologyCreateNode{
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

	go func() {
		timer := time.NewTimer(s.config.Assessment.IncrementalUpdateDelay)
		defer timer.Stop()

		<-timer.C

		ctx, cancel := context.WithTimeout(context.Background(), s.config.Assessment.IncrementalUpdateTimeout)
		defer cancel()

		if err := s.assessmentService.UpdateSkillProfileForDimension(ctx, userID, dimension); err != nil {
			s.logger.Error("更新能力画像失败",
				zap.Int64("user_id", userID),
				zap.String("dimension", dimension),
				zap.Error(err))
		}
	}()
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

func toInstanceInfo(inst *model.Instance) *dto.InstanceInfo {
	remaining := int64(time.Until(inst.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return &dto.InstanceInfo{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		Status:           inst.Status,
		AccessURL:        inst.AccessURL,
		ExpiresAt:        inst.ExpiresAt,
		RemainingTime:    remaining,
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
