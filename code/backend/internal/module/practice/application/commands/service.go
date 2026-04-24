package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/module/practice/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
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

type Service struct {
	repo              practiceports.PracticeCommandRepository
	challengeRepo     challengecontracts.PracticeChallengeContract
	imageRepo         challengecontracts.ImageStore
	instanceRepo      practiceports.InstanceRepository
	runtimeService    practiceports.RuntimeInstanceService
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
	repo practiceports.PracticeCommandRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
	imageRepo challengecontracts.ImageStore,
	instanceRepo practiceports.InstanceRepository,
	runtimeService practiceports.RuntimeInstanceService,
	scoreService ScoreUpdater,
	assessmentService AssessmentService,
	redis *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg == nil {
		cfg = &config.Config{}
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
		config:            cfg,
		logger:            logger,
		baseCtx:           baseCtx,
		cancel:            cancel,
	}
}

func (s *Service) StartChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.startPersonalChallenge(ctx, userID, challengeID)
}

func (s *Service) StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*dto.InstanceResp, error) {
	scope, err := s.resolveContestChallengeInstanceScope(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, err
	}
	return s.startChallengeWithScope(ctx, userID, challengeID, scope)
}

func (s *Service) StartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*dto.InstanceResp, error) {
	challengeID, scope, err := s.resolveContestAWDServiceInstanceScope(ctx, userID, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	return s.startChallengeWithScope(ctx, userID, challengeID, scope)
}

func (s *Service) startPersonalChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error) {
	return s.startChallengeWithScope(ctx, userID, challengeID, practiceports.InstanceScope{
		FlagSubjectID: userID,
		ShareScope:    model.InstanceSharingPerUser,
	})
}

func (s *Service) startChallengeWithScope(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*dto.InstanceResp, error) {
	chal, topology, err := s.loadRuntimeSubjectWithScope(ctx, scope, challengeID)
	if err != nil {
		return nil, err
	}
	if chal.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}
	if chal.ImageID == 0 {
		if topology == nil {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New(errMsgChallengeNoTarget))
		}
	}
	scope = resolveEffectiveInstanceScope(chal, scope)

	flag, nonce, err := s.buildInstanceFlag(scope.FlagSubjectID, challengeID, chal)
	if err != nil {
		return nil, err
	}

	var (
		instance *model.Instance
		reused   bool
	)
	initialStatus := model.InstanceStatusCreating
	if s.schedulerEnabled() {
		initialStatus = model.InstanceStatusPending
	}
	if err := s.repo.WithinTransaction(ctx, func(txRepo practiceports.PracticeCommandTxRepository) error {
		if err := txRepo.LockInstanceScopeWithContext(ctx, userID, challengeID, scope); err != nil {
			return err
		}

		existingInstance, err := txRepo.FindScopedExistingInstanceWithContext(ctx, userID, challengeID, scope)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if existingInstance != nil {
			if scope.ShareScope == model.InstanceSharingShared {
				refreshedExpiry := existingInstance.ExpiresAt
				candidateExpiry := time.Now().Add(s.config.Container.DefaultTTL)
				if candidateExpiry.After(refreshedExpiry) {
					refreshedExpiry = candidateExpiry
				}
				if err := txRepo.RefreshInstanceExpiryWithContext(ctx, existingInstance.ID, refreshedExpiry); err != nil {
					return errcode.ErrInternal.WithCause(err)
				}
				existingInstance.ExpiresAt = refreshedExpiry
			}
			instance = existingInstance
			reused = true
			return nil
		}

		runningCount, err := txRepo.CountScopedRunningInstancesWithContext(ctx, userID, scope)
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

		hostPort, err := txRepo.ReserveAvailablePortWithContext(ctx, s.config.Container.PortRangeStart, s.config.Container.PortRangeEnd)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}

		instance = &model.Instance{
			UserID:      userID,
			ContestID:   scope.ContestID,
			TeamID:      scope.TeamID,
			ChallengeID: challengeID,
			ServiceID:   scope.ServiceID,
			HostPort:    hostPort,
			ShareScope:  scope.ShareScope,
			Status:      initialStatus,
			Nonce:       nonce,
			ExpiresAt:   time.Now().Add(s.config.Container.DefaultTTL),
			MaxExtends:  s.config.Container.MaxExtends,
		}
		if err := txRepo.CreateInstanceWithContext(ctx, instance); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if err := txRepo.BindReservedPortWithContext(ctx, hostPort, instance.ID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if reused {
		return domain.InstanceRespFromModel(instance), nil
	}
	if s.schedulerEnabled() {
		s.logger.Info("实例已入启动队列",
			zap.Int64("user_id", userID),
			zap.Int64("challenge_id", challengeID),
			zap.Int64("instance_id", instance.ID))
		return domain.InstanceRespFromModel(instance), nil
	}

	if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
		return nil, err
	}
	return domain.InstanceRespFromModel(instance), nil
}

func (s *Service) markInstanceFailed(ctx context.Context, instance *model.Instance) {
	if instance == nil {
		return
	}
	if err := s.runtimeService.CleanupRuntime(instance); err != nil {
		s.logger.Warn("清理失败实例运行时资源失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.instanceRepo.UpdateStatusAndReleasePortWithContext(ctx, instance.ID, model.InstanceStatusFailed); err != nil {
		s.logger.Warn("更新失败实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
	}
}

func (s *Service) RunProvisioningLoop(ctx context.Context) {
	if !s.schedulerEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ticker := time.NewTicker(s.schedulerPollInterval())
	defer ticker.Stop()

	for {
		if err := s.dispatchPendingInstances(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("调度待启动实例失败", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) dispatchPendingInstances(ctx context.Context) error {
	limit, err := s.availableProvisioningSlots(ctx)
	if err != nil {
		return err
	}
	if limit <= 0 {
		return nil
	}

	instances, err := s.instanceRepo.ListPendingInstancesWithContext(ctx, limit)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		claimed, err := s.instanceRepo.TryTransitionStatusWithContext(ctx, instance.ID, model.InstanceStatusPending, model.InstanceStatusCreating)
		if err != nil {
			return err
		}
		if !claimed {
			continue
		}

		instanceID := instance.ID
		s.runAsyncTask(func(taskCtx context.Context) {
			s.processPendingInstance(taskCtx, instanceID)
		})
	}
	return nil
}

func (s *Service) availableProvisioningSlots(ctx context.Context) (int, error) {
	slots := s.schedulerMaxConcurrentStarts()
	if slots <= 0 {
		return 0, nil
	}

	creatingCount, err := s.instanceRepo.CountInstancesByStatusWithContext(ctx, []string{model.InstanceStatusCreating})
	if err != nil {
		return 0, err
	}
	slots -= int(creatingCount)
	if slots <= 0 {
		return 0, nil
	}

	maxActive := s.schedulerMaxActiveInstances()
	if maxActive > 0 {
		activeCount, err := s.instanceRepo.CountInstancesByStatusWithContext(ctx, []string{model.InstanceStatusCreating, model.InstanceStatusRunning})
		if err != nil {
			return 0, err
		}
		remainingCapacity := maxActive - int(activeCount)
		if remainingCapacity <= 0 {
			return 0, nil
		}
		if remainingCapacity < slots {
			slots = remainingCapacity
		}
	}

	batchSize := s.schedulerBatchSize()
	if batchSize > 0 && batchSize < slots {
		slots = batchSize
	}
	return slots, nil
}

func (s *Service) processPendingInstance(ctx context.Context, instanceID int64) {
	instance, err := s.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		s.logger.Error("读取待启动实例失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		return
	}
	if instance == nil || instance.Status != model.InstanceStatusCreating {
		return
	}

	chal, topology, err := s.loadRuntimeSubjectForInstance(ctx, instance)
	if err != nil {
		s.logger.Error("读取题目失败", zap.Int64("instance_id", instanceID), zap.Int64("challenge_id", instance.ChallengeID), zap.Error(err))
		s.markInstanceFailed(ctx, instance)
		return
	}

	flag, err := s.buildProvisioningFlag(instance, chal)
	if err != nil {
		s.logger.Error("生成实例 Flag 失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		s.markInstanceFailed(ctx, instance)
		return
	}

	if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
		s.logger.Warn("实例异步启动失败", zap.Int64("instance_id", instanceID), zap.Error(err))
	}
}

func (s *Service) provisionInstance(ctx context.Context, instance *model.Instance, chal *model.Challenge, topology *model.ChallengeTopology, flag string) error {
	createCtx, cancel := context.WithTimeout(ctx, s.config.Container.CreateTimeout)
	defer cancel()

	if err := s.createContainer(createCtx, instance, chal, topology, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		s.markInstanceFailed(ctx, instance)
		return err
	}
	if err := s.waitForInstanceReadiness(createCtx, instance.AccessURL); err != nil {
		s.logger.Error("实例访问地址未就绪", zap.Error(err), zap.Int64("instance_id", instance.ID), zap.String("access_url", instance.AccessURL))
		s.markInstanceFailed(ctx, instance)
		return errcode.ErrContainerStartFailed.WithCause(err)
	}

	instance.Status = model.InstanceStatusRunning
	if err := s.instanceRepo.UpdateRuntimeWithContext(ctx, instance); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		s.markInstanceFailed(ctx, instance)
		return errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("实例启动成功",
		zap.Int64("user_id", instance.UserID),
		zap.Int64("challenge_id", instance.ChallengeID),
		zap.Int64("instance_id", instance.ID))
	return nil
}

func (s *Service) waitForInstanceReadiness(ctx context.Context, accessURL string) error {
	if strings.TrimSpace(accessURL) == "" {
		return fmt.Errorf("instance access url is empty")
	}

	attempts := s.startProbeAttempts()
	client := &http.Client{Timeout: s.startProbeTimeout()}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		lastErr = s.probeInstanceAccessURL(ctx, client, accessURL)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt == attempts-1 {
			break
		}

		timer := time.NewTimer(s.startProbeInterval())
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func (s *Service) probeInstanceAccessURL(ctx context.Context, client *http.Client, accessURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
	return nil
}

func (s *Service) buildProvisioningFlag(instance *model.Instance, chal *model.Challenge) (string, error) {
	if instance == nil || chal == nil {
		return "", errcode.ErrInternal.WithCause(fmt.Errorf("instance or challenge is nil"))
	}

	switch chal.FlagType {
	case model.FlagTypeDynamic:
		if strings.TrimSpace(instance.Nonce) == "" {
			return "", errcode.ErrInternal.WithCause(fmt.Errorf("instance nonce is empty"))
		}
		if strings.TrimSpace(s.config.Container.FlagGlobalSecret) == "" {
			return "", errcode.ErrInternal.WithCause(fmt.Errorf("flag global secret is empty"))
		}
		subjectID := instance.UserID
		if instance.TeamID != nil && *instance.TeamID > 0 {
			subjectID = *instance.TeamID
		}
		return crypto.GenerateDynamicFlag(subjectID, chal.ID, s.config.Container.FlagGlobalSecret, instance.Nonce, chal.FlagPrefix), nil
	case model.FlagTypeStatic:
		return chal.FlagHash, nil
	default:
		return "", nil
	}
}

func (s *Service) schedulerEnabled() bool {
	return s != nil && s.config != nil && s.config.Container.Scheduler.Enabled
}

func (s *Service) schedulerPollInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.PollInterval <= 0 {
		return time.Second
	}
	return s.config.Container.Scheduler.PollInterval
}

func (s *Service) schedulerBatchSize() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.BatchSize <= 0 {
		return 1
	}
	return s.config.Container.Scheduler.BatchSize
}

func (s *Service) schedulerMaxConcurrentStarts() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.MaxConcurrentStarts <= 0 {
		return 1
	}
	return s.config.Container.Scheduler.MaxConcurrentStarts
}

func (s *Service) schedulerMaxActiveInstances() int {
	if s == nil || s.config == nil {
		return 0
	}
	return s.config.Container.Scheduler.MaxActiveInstances
}

func (s *Service) startProbeTimeout() time.Duration {
	if s == nil || s.config == nil || s.config.Container.StartProbeTimeout <= 0 {
		return 800 * time.Millisecond
	}
	return s.config.Container.StartProbeTimeout
}

func (s *Service) startProbeInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.StartProbeInterval <= 0 {
		return 300 * time.Millisecond
	}
	return s.config.Container.StartProbeInterval
}

func (s *Service) startProbeAttempts() int {
	if s == nil || s.config == nil || s.config.Container.StartProbeAttempts <= 0 {
		return 5
	}
	return s.config.Container.StartProbeAttempts
}

func (s *Service) SubmitFlag(ctx context.Context, userID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	challengeItem, err := s.challengeRepo.FindByID(ctx, challengeID)
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

	alreadySolved := false
	if _, err := s.repo.FindCorrectSubmissionWithContext(ctx, userID, challengeID); err == nil {
		alreadySolved = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if alreadySolved && challengeItem.FlagType == model.FlagTypeManualReview {
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

	submission := &model.Submission{
		UserID:       userID,
		ChallengeID:  challengeID,
		Flag:         "",
		IsCorrect:    false,
		ReviewStatus: model.SubmissionReviewStatusNotRequired,
		SubmittedAt:  time.Now(),
		UpdatedAt:    time.Now(),
	}
	status := dto.SubmissionStatusIncorrect
	submissionPersisted := false

	if challengeItem.FlagType == model.FlagTypeManualReview {
		submission.Flag = flag
		submission.ReviewStatus = model.SubmissionReviewStatusPending
		status = dto.SubmissionStatusPendingReview
	} else {
		isCorrect, err := s.validateSubmittedFlag(ctx, userID, challengeItem, flag)
		if err != nil {
			return nil, err
		}
		submission.IsCorrect = isCorrect
		if isCorrect {
			status = dto.SubmissionStatusCorrect
			if alreadySolved {
				auditlog.MarkSkip(ctx)
				return &dto.SubmissionResp{
					IsCorrect:   true,
					Status:      status,
					SubmittedAt: submission.SubmittedAt,
				}, nil
			}
		}
	}

	if !submissionPersisted {
		if err := s.repo.CreateSubmissionWithContext(ctx, submission); err != nil {
			if submission.IsCorrect && s.repo.IsUniqueViolation(err) {
				return nil, errcode.ErrAlreadySolved
			}
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	if submission.IsCorrect && !alreadySolved {
		cacheKey := constants.UserProgressKey(userID)
		if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
			s.logger.Warn("删除进度缓存失败", zap.Int64("user_id", userID), zap.Error(err))
		}
		s.publishWeakEvent(ctx, platformevents.Event{
			Name: practicecontracts.EventFlagAccepted,
			Payload: practicecontracts.FlagAcceptedEvent{
				UserID:      userID,
				ChallengeID: challengeID,
				Dimension:   challengeItem.Category,
				Points:      challengeItem.Points,
				OccurredAt:  submission.SubmittedAt,
			},
		})
	}

	var instanceShutdownAt *time.Time
	if submission.IsCorrect && !alreadySolved {
		instanceShutdownAt = s.applySolveGracePeriod(ctx, userID, challengeItem, submission.SubmittedAt)
	}

	resp := &dto.SubmissionResp{
		IsCorrect:          submission.IsCorrect,
		Status:             status,
		SubmittedAt:        submission.SubmittedAt,
		InstanceShutdownAt: instanceShutdownAt,
	}
	if submission.IsCorrect && !alreadySolved {
		resp.Points = challengeItem.Points
		if s.scoreService != nil {
			s.triggerScoreUpdate(userID)
		}
	}

	return resp, nil
}

func (s *Service) applySolveGracePeriod(ctx context.Context, userID int64, challengeItem *model.Challenge, solvedAt time.Time) *time.Time {
	if s == nil || s.instanceRepo == nil || challengeItem == nil {
		return nil
	}

	gracePeriod := s.config.Container.SolveGracePeriod
	if gracePeriod <= 0 {
		return nil
	}

	instance, err := s.instanceRepo.FindByUserAndChallengeWithContext(ctx, userID, challengeItem.ID)
	if err != nil {
		s.logger.Warn("查询解题后实例失败", zap.Int64("user_id", userID), zap.Int64("challenge_id", challengeItem.ID), zap.Error(err))
		return nil
	}
	if instance == nil || instance.ShareScope != model.InstanceSharingPerUser {
		return nil
	}

	shutdownAt := instance.ExpiresAt
	graceExpiry := solvedAt.Add(gracePeriod)
	if shutdownAt.After(graceExpiry) {
		shutdownAt = graceExpiry
		if err := s.instanceRepo.RefreshInstanceExpiryWithContext(ctx, instance.ID, shutdownAt); err != nil {
			s.logger.Warn("收缩解题后实例生命周期失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
			return nil
		}
	}

	return &shutdownAt
}

func formatSolveGracePeriod(delay time.Duration) string {
	if delay <= 0 || delay < time.Minute {
		return "1 分钟内"
	}
	if delay%time.Hour == 0 {
		return fmt.Sprintf("%d 小时", int(delay/time.Hour))
	}

	minutes := int(delay.Round(time.Minute) / time.Minute)
	if minutes <= 1 {
		return "1 分钟"
	}
	return fmt.Sprintf("%d 分钟", minutes)
}

func (s *Service) ReviewManualReviewSubmissionWithContext(
	ctx context.Context,
	submissionID, reviewerID int64,
	reviewerRole string,
	req *dto.ReviewManualReviewSubmissionReq,
) (*dto.TeacherManualReviewSubmissionDetailResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ensureManualReviewRequesterRole(reviewerRole); err != nil {
		return nil, err
	}
	if err := ensureManualReviewDecisionStatus(req); err != nil {
		return nil, err
	}
	record, err := s.repo.GetTeacherManualReviewSubmissionByIDWithContext(ctx, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessManualReviewSubmission(ctx, s.repo, reviewerID, reviewerRole, record); err != nil {
		return nil, err
	}
	if record.Submission.ReviewStatus != model.SubmissionReviewStatusPending {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("仅待审核提交可执行评阅"))
	}

	challengeItem, err := s.challengeRepo.FindByID(ctx, record.Submission.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.FlagType != model.FlagTypeManualReview {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("当前提交不属于人工审核题"))
	}

	now := time.Now()
	item := record.Submission
	item.ReviewStatus = req.ReviewStatus
	item.ReviewComment = strings.TrimSpace(req.ReviewComment)
	item.ReviewedBy = &reviewerID
	item.ReviewedAt = &now
	item.UpdatedAt = now
	if req.ReviewStatus == model.SubmissionReviewStatusApproved {
		if _, err := s.repo.FindCorrectSubmissionWithContext(ctx, item.UserID, item.ChallengeID); err == nil {
			return nil, errcode.ErrAlreadySolved
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		item.IsCorrect = true
		item.Score = challengeItem.Points
	} else {
		item.IsCorrect = false
		item.Score = 0
	}

	if err := s.repo.UpdateSubmissionWithContext(ctx, &item); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if item.IsCorrect {
		if s.redis != nil {
			if err := s.redis.Del(ctx, constants.UserProgressKey(item.UserID)).Err(); err != nil {
				s.logger.Warn("删除进度缓存失败", zap.Int64("user_id", item.UserID), zap.Error(err))
			}
		}
		s.publishWeakEvent(ctx, platformevents.Event{
			Name: practicecontracts.EventFlagAccepted,
			Payload: practicecontracts.FlagAcceptedEvent{
				UserID:      item.UserID,
				ChallengeID: item.ChallengeID,
				Dimension:   challengeItem.Category,
				Points:      item.Score,
				OccurredAt:  now,
			},
		})
		if s.scoreService != nil {
			s.triggerScoreUpdate(item.UserID)
		}
	}

	return manualReviewDetailRespFromRecord(*record, item), nil
}

func (s *Service) ListTeacherManualReviewSubmissions(
	requesterID int64,
	requesterRole string,
	query *dto.TeacherManualReviewSubmissionQuery,
) (*dto.PageResult, error) {
	return s.ListTeacherManualReviewSubmissionsWithContext(context.Background(), requesterID, requesterRole, query)
}

func (s *Service) ListTeacherManualReviewSubmissionsWithContext(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherManualReviewSubmissionQuery,
) (*dto.PageResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	if query == nil {
		query = &dto.TeacherManualReviewSubmissionQuery{}
	}
	normalized, err := normalizeTeacherManualReviewQueryWithContext(ctx, s.repo, requesterID, requesterRole, query)
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListTeacherManualReviewSubmissionsWithContext(ctx, normalized)
	if err != nil {
		return nil, err
	}

	respItems := make([]*dto.TeacherManualReviewSubmissionItemResp, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, manualReviewListItemRespFromRecord(item))
	}

	return &dto.PageResult{
		List:  respItems,
		Total: total,
		Page:  normalized.Page,
		Size:  normalized.Size,
	}, nil
}

func (s *Service) GetTeacherManualReviewSubmission(
	submissionID, requesterID int64,
	requesterRole string,
) (*dto.TeacherManualReviewSubmissionDetailResp, error) {
	return s.GetTeacherManualReviewSubmissionWithContext(context.Background(), submissionID, requesterID, requesterRole)
}

func (s *Service) GetTeacherManualReviewSubmissionWithContext(
	ctx context.Context,
	submissionID, requesterID int64,
	requesterRole string,
) (*dto.TeacherManualReviewSubmissionDetailResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	record, err := s.repo.GetTeacherManualReviewSubmissionByIDWithContext(ctx, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessManualReviewSubmission(ctx, s.repo, requesterID, requesterRole, record); err != nil {
		return nil, err
	}
	return manualReviewDetailRespFromRecord(*record, record.Submission), nil
}

func (s *Service) ListMyChallengeSubmissions(userID, challengeID int64) ([]*dto.ChallengeSubmissionRecordResp, error) {
	return s.ListMyChallengeSubmissionsWithContext(context.Background(), userID, challengeID)
}

func (s *Service) ListMyChallengeSubmissionsWithContext(ctx context.Context, userID, challengeID int64) ([]*dto.ChallengeSubmissionRecordResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	challengeItem, err := s.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	items, err := s.repo.ListChallengeSubmissionsWithContext(ctx, userID, challengeID, 20)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*dto.ChallengeSubmissionRecordResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, challengeSubmissionRecordRespFromModel(item))
	}
	return resp, nil
}

func ensureTeacherCanAccessManualReviewSubmission(
	ctx context.Context,
	repo practiceports.PracticeCommandRepository,
	requesterID int64,
	requesterRole string,
	record *practiceports.TeacherManualReviewSubmissionRecord,
) error {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return err
	}
	if requesterRole == model.RoleAdmin {
		return nil
	}
	requester, err := repo.FindUserByIDWithContext(ctx, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUnauthorized
		}
		return err
	}
	if requester.ClassName == "" || requester.ClassName != record.ClassName {
		return errcode.ErrForbidden
	}
	return nil
}

func normalizeTeacherManualReviewQuery(
	repo practiceports.PracticeCommandRepository,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherManualReviewSubmissionQuery,
) (*dto.TeacherManualReviewSubmissionQuery, error) {
	return normalizeTeacherManualReviewQueryWithContext(context.Background(), repo, requesterID, requesterRole, query)
}

func normalizeTeacherManualReviewQueryWithContext(
	ctx context.Context,
	repo practiceports.PracticeCommandRepository,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherManualReviewSubmissionQuery,
) (*dto.TeacherManualReviewSubmissionQuery, error) {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	if err := ensureManualReviewQuery(query); err != nil {
		return nil, err
	}
	normalized := *query
	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.Size <= 0 {
		normalized.Size = 20
	}
	if requesterRole == model.RoleAdmin {
		return &normalized, nil
	}

	requester, err := repo.FindUserByIDWithContext(ctx, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUnauthorized
		}
		return nil, err
	}
	if requester.ClassName == "" {
		return nil, errcode.ErrForbidden
	}
	if normalized.ClassName != "" && normalized.ClassName != requester.ClassName {
		return nil, errcode.ErrForbidden
	}
	normalized.ClassName = requester.ClassName
	return &normalized, nil
}

func ensureManualReviewRequesterRole(role string) error {
	if role == model.RoleAdmin || role == model.RoleTeacher {
		return nil
	}
	return errcode.ErrForbidden
}

func ensureManualReviewDecisionStatus(req *dto.ReviewManualReviewSubmissionReq) error {
	if req == nil {
		return errcode.ErrInvalidParams.WithCause(errors.New("评阅请求不能为空"))
	}
	if len([]rune(strings.TrimSpace(req.ReviewComment))) > 4000 {
		return errcode.ErrInvalidParams.WithCause(errors.New("review_comment 不能超过 4000 个字符"))
	}
	if req.ReviewStatus == model.SubmissionReviewStatusApproved || req.ReviewStatus == model.SubmissionReviewStatusRejected {
		return nil
	}
	return errcode.ErrInvalidParams.WithCause(errors.New("review_status 仅支持 approved 或 rejected"))
}

func ensureManualReviewQuery(query *dto.TeacherManualReviewSubmissionQuery) error {
	if query == nil {
		return nil
	}
	if query.StudentID != nil && *query.StudentID <= 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New("student_id 必须大于 0"))
	}
	if query.ChallengeID != nil && *query.ChallengeID <= 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New("challenge_id 必须大于 0"))
	}
	if len([]rune(strings.TrimSpace(query.ClassName))) > 128 {
		return errcode.ErrInvalidParams.WithCause(errors.New("class_name 不能超过 128 个字符"))
	}
	if query.Size > 100 {
		return errcode.ErrInvalidParams.WithCause(errors.New("page_size 不能超过 100"))
	}
	if query.ReviewStatus == "" ||
		query.ReviewStatus == model.SubmissionReviewStatusPending ||
		query.ReviewStatus == model.SubmissionReviewStatusApproved ||
		query.ReviewStatus == model.SubmissionReviewStatusRejected {
		return nil
	}
	return errcode.ErrInvalidParams.WithCause(errors.New("review_status 仅支持 pending、approved 或 rejected"))
}

func manualReviewDetailRespFromRecord(
	record practiceports.TeacherManualReviewSubmissionRecord,
	submission model.Submission,
) *dto.TeacherManualReviewSubmissionDetailResp {
	return &dto.TeacherManualReviewSubmissionDetailResp{
		ID:              submission.ID,
		UserID:          submission.UserID,
		StudentUsername: record.StudentUsername,
		StudentName:     record.StudentName,
		ClassName:       record.ClassName,
		ChallengeID:     submission.ChallengeID,
		ChallengeTitle:  record.ChallengeTitle,
		Answer:          submission.Flag,
		IsCorrect:       submission.IsCorrect,
		Score:           submission.Score,
		ReviewStatus:    submission.ReviewStatus,
		ReviewedBy:      submission.ReviewedBy,
		ReviewedAt:      submission.ReviewedAt,
		ReviewComment:   submission.ReviewComment,
		SubmittedAt:     submission.SubmittedAt,
		UpdatedAt:       submission.UpdatedAt,
		ReviewerName:    record.ReviewerName,
	}
}

func manualReviewListItemRespFromRecord(record practiceports.TeacherManualReviewSubmissionRecord) *dto.TeacherManualReviewSubmissionItemResp {
	answerPreview := strings.TrimSpace(record.Submission.Flag)
	if len([]rune(answerPreview)) > 80 {
		answerPreview = string([]rune(answerPreview)[:80]) + "..."
	}
	return &dto.TeacherManualReviewSubmissionItemResp{
		ID:              record.Submission.ID,
		UserID:          record.Submission.UserID,
		StudentUsername: record.StudentUsername,
		StudentName:     record.StudentName,
		ClassName:       record.ClassName,
		ChallengeID:     record.Submission.ChallengeID,
		ChallengeTitle:  record.ChallengeTitle,
		AnswerPreview:   answerPreview,
		ReviewStatus:    record.Submission.ReviewStatus,
		SubmittedAt:     record.Submission.SubmittedAt,
		ReviewedAt:      record.Submission.ReviewedAt,
		UpdatedAt:       record.Submission.UpdatedAt,
	}
}

func challengeSubmissionRecordRespFromModel(item model.Submission) *dto.ChallengeSubmissionRecordResp {
	status := dto.SubmissionStatusIncorrect
	answer := ""

	if item.ReviewStatus == model.SubmissionReviewStatusPending {
		status = dto.SubmissionStatusPendingReview
		answer = item.Flag
	} else if item.IsCorrect {
		status = dto.SubmissionStatusCorrect
	}

	return &dto.ChallengeSubmissionRecordResp{
		ID:          item.ID,
		Status:      status,
		Answer:      answer,
		SubmittedAt: item.SubmittedAt,
	}
}

func (s *Service) resolveContestChallengeInstanceScope(ctx context.Context, userID, contestID, challengeID int64) (practiceports.InstanceScope, error) {
	scope, err := s.resolveContestBaseInstanceScope(ctx, userID, contestID)
	if err != nil {
		return practiceports.InstanceScope{}, err
	}
	if scope.ContestMode == model.ContestModeAWD {
		return practiceports.InstanceScope{}, errcode.ErrInvalidParams.WithCause(
			errors.New("awd 赛事实例启动必须使用 service_id 入口"),
		)
	}
	contestChallenge, err := s.repo.FindContestChallengeWithContext(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if !contestChallenge.IsVisible {
		return practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}
	return scope, nil
}

func (s *Service) resolveContestAWDServiceInstanceScope(ctx context.Context, userID, contestID, serviceID int64) (int64, practiceports.InstanceScope, error) {
	scope, err := s.resolveContestBaseInstanceScope(ctx, userID, contestID)
	if err != nil {
		return 0, practiceports.InstanceScope{}, err
	}
	service, err := s.repo.FindContestAWDServiceWithContext(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, practiceports.InstanceScope{}, errcode.ErrChallengeNotInContest
		}
		return 0, practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	if !service.IsVisible {
		return 0, practiceports.InstanceScope{}, errcode.ErrContestChallengeVisible
	}
	serviceIDCopy := service.ID
	scope.ServiceID = &serviceIDCopy
	return service.ChallengeID, scope, nil
}

func (s *Service) loadRuntimeSubjectWithScope(ctx context.Context, scope practiceports.InstanceScope, challengeID int64) (*model.Challenge, *model.ChallengeTopology, error) {
	if scope.ServiceID != nil && scope.ContestID != nil {
		return s.loadContestAWDServiceRuntimeSubject(ctx, *scope.ContestID, *scope.ServiceID)
	}

	chal, err := s.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errcode.ErrChallengeNotFound
		}
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	topology, err := s.challengeRepo.FindChallengeTopologyByChallengeID(ctx, chal.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errcode.ErrContainerCreateFailed.WithCause(err)
	}
	return chal, topology, nil
}

func (s *Service) loadRuntimeSubjectForInstance(ctx context.Context, instance *model.Instance) (*model.Challenge, *model.ChallengeTopology, error) {
	if instance != nil && instance.ServiceID != nil && instance.ContestID != nil {
		return s.loadContestAWDServiceRuntimeSubject(ctx, *instance.ContestID, *instance.ServiceID)
	}
	return s.loadRuntimeSubjectWithScope(ctx, practiceports.InstanceScope{}, instance.ChallengeID)
}

func (s *Service) loadContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*model.Challenge, *model.ChallengeTopology, error) {
	service, err := s.repo.FindContestAWDServiceWithContext(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.ErrChallengeNotInContest
		}
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	snapshot, err := model.DecodeContestAWDServiceSnapshot(service.ServiceSnapshot)
	if err != nil {
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	chal := buildContestAWDServiceVirtualChallenge(service, snapshot)
	topology, err := buildContestAWDServiceVirtualTopology(service, snapshot)
	if err != nil {
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}
	return chal, topology, nil
}

func (s *Service) resolveContestBaseInstanceScope(ctx context.Context, userID, contestID int64) (practiceports.InstanceScope, error) {
	if s.repo == nil {
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(fmt.Errorf("practice repository is nil"))
	}

	contest, err := s.repo.FindContestByIDWithContext(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrContestNotFound
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch contest.Status {
	case model.ContestStatusRunning, model.ContestStatusFrozen:
	default:
		if contest.Status == model.ContestStatusEnded {
			return practiceports.InstanceScope{}, errcode.ErrContestEnded
		}
		return practiceports.InstanceScope{}, errcode.ErrContestNotRunning
	}

	registration, err := s.repo.FindContestRegistrationWithContext(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return practiceports.InstanceScope{}, errcode.ErrNotRegistered
		}
		return practiceports.InstanceScope{}, errcode.ErrInternal.WithCause(err)
	}
	switch registration.Status {
	case model.ContestRegistrationStatusApproved:
	case model.ContestRegistrationStatusPending:
		return practiceports.InstanceScope{}, errcode.ErrContestRegistrationPending
	default:
		return practiceports.InstanceScope{}, errcode.ErrRegistrationNotApproved
	}

	contestIDCopy := contestID
	scope := practiceports.InstanceScope{
		ContestID:     &contestIDCopy,
		ContestMode:   contest.Mode,
		FlagSubjectID: userID,
		ShareScope:    model.InstanceSharingPerUser,
	}
	if registration.TeamID != nil && *registration.TeamID > 0 {
		teamID := *registration.TeamID
		scope.TeamID = &teamID
	}

	return scope, nil
}

func resolveEffectiveInstanceScope(chal *model.Challenge, scope practiceports.InstanceScope) practiceports.InstanceScope {
	effective := scope
	effective.FlagSubjectID = scope.FlagSubjectID
	effective.ShareScope = model.InstanceSharingPerUser

	switch {
	case scope.ContestMode == model.ContestModeAWD:
		effective.ShareScope = model.InstanceSharingPerTeam
		if scope.TeamID != nil && *scope.TeamID > 0 {
			effective.FlagSubjectID = *scope.TeamID
		}
	case chal.InstanceSharing == model.InstanceSharingShared:
		effective.ShareScope = model.InstanceSharingShared
		effective.TeamID = nil
	case chal.InstanceSharing == model.InstanceSharingPerTeam && scope.TeamID != nil && *scope.TeamID > 0:
		effective.ShareScope = model.InstanceSharingPerTeam
		effective.FlagSubjectID = *scope.TeamID
	default:
		effective.ShareScope = model.InstanceSharingPerUser
		effective.TeamID = nil
	}

	if effective.ShareScope != model.InstanceSharingPerTeam {
		effective.TeamID = nil
	}
	return effective
}

func buildContestAWDServiceVirtualChallenge(service *model.ContestAWDService, snapshot model.ContestAWDServiceSnapshot) *model.Challenge {
	chal := &model.Challenge{
		ID:              service.ChallengeID,
		Title:           firstRuntimeValue(service.DisplayName, snapshot.Name),
		Category:        snapshot.Category,
		Difficulty:      snapshot.Difficulty,
		Points:          parseContestAWDServiceSnapshotPoints(service.ScoreConfig),
		Status:          model.ChallengeStatusPublished,
		ImageID:         parseContestAWDServiceSnapshotImageID(snapshot.RuntimeConfig),
		FlagType:        parseContestAWDServiceSnapshotFlagType(snapshot.FlagConfig),
		FlagPrefix:      parseContestAWDServiceSnapshotFlagPrefix(snapshot.FlagConfig),
		InstanceSharing: parseContestAWDServiceSnapshotInstanceSharing(snapshot.RuntimeConfig),
	}
	if chal.FlagPrefix == "" {
		chal.FlagPrefix = "flag"
	}
	return chal
}

func buildContestAWDServiceVirtualTopology(service *model.ContestAWDService, snapshot model.ContestAWDServiceSnapshot) (*model.ChallengeTopology, error) {
	topologyPayload, ok := snapshot.RuntimeConfig["topology"]
	if !ok {
		return nil, nil
	}
	topologyMap, ok := topologyPayload.(map[string]any)
	if !ok {
		return nil, nil
	}
	entryNodeKey, _ := topologyMap["entry_node_key"].(string)
	specPayload, ok := topologyMap["spec"]
	if !ok {
		return nil, nil
	}
	specRaw, err := json.Marshal(specPayload)
	if err != nil {
		return nil, err
	}
	return &model.ChallengeTopology{
		ChallengeID:  service.ChallengeID,
		EntryNodeKey: strings.TrimSpace(entryNodeKey),
		Spec:         string(specRaw),
	}, nil
}

func parseContestAWDServiceSnapshotPoints(scoreConfig string) int {
	if scoreConfig == "" {
		return 0
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(scoreConfig), &payload); err != nil {
		return 0
	}
	return parseContestAWDServiceSnapshotInt(payload["points"])
}

func parseContestAWDServiceSnapshotImageID(runtimeConfig map[string]any) int64 {
	if runtimeConfig == nil {
		return 0
	}
	value := parseContestAWDServiceSnapshotInt(runtimeConfig["image_id"])
	if value <= 0 {
		return 0
	}
	return int64(value)
}

func parseContestAWDServiceSnapshotInstanceSharing(runtimeConfig map[string]any) model.InstanceSharing {
	if runtimeConfig == nil {
		return model.InstanceSharingPerTeam
	}
	value, _ := runtimeConfig["instance_sharing"].(string)
	switch model.InstanceSharing(value) {
	case model.InstanceSharingShared:
		return model.InstanceSharingShared
	case model.InstanceSharingPerUser:
		return model.InstanceSharingPerUser
	case model.InstanceSharingPerTeam:
		return model.InstanceSharingPerTeam
	default:
		return model.InstanceSharingPerTeam
	}
}

func parseContestAWDServiceSnapshotFlagType(flagConfig map[string]any) string {
	if flagConfig == nil {
		return model.FlagTypeDynamic
	}
	value, _ := flagConfig["flag_type"].(string)
	if strings.TrimSpace(value) == "" {
		return model.FlagTypeDynamic
	}
	return value
}

func parseContestAWDServiceSnapshotFlagPrefix(flagConfig map[string]any) string {
	if flagConfig == nil {
		return "flag"
	}
	value, _ := flagConfig["flag_prefix"].(string)
	if strings.TrimSpace(value) == "" {
		return "flag"
	}
	return value
}

func parseContestAWDServiceSnapshotInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		next, err := typed.Int64()
		if err != nil {
			return 0
		}
		return int(next)
	default:
		return 0
	}
}

func firstRuntimeValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
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

func (s *Service) validateSubmittedFlag(ctx context.Context, userID int64, challengeItem *model.Challenge, flag string) (bool, error) {
	switch challengeItem.FlagType {
	case model.FlagTypeStatic:
		inputHash := crypto.HashStaticFlag(flag, challengeItem.FlagSalt)
		return crypto.ValidateFlag(inputHash, challengeItem.FlagHash), nil
	case model.FlagTypeRegex:
		return regexp.MatchString(challengeItem.FlagRegex, flag)
	case model.FlagTypeManualReview:
		return false, nil
	case model.FlagTypeDynamic:
	default:
		return false, errcode.ErrInvalidParams.WithCause(fmt.Errorf("unsupported flag type %s", challengeItem.FlagType))
	}

	instance, err := s.instanceRepo.FindByUserAndChallengeWithContext(ctx, userID, challengeItem.ID)
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

func (s *Service) createContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, topology *model.ChallengeTopology, flag string) error {
	if topology == nil {
		return s.createSingleContainer(ctx, instance, chal, flag)
	}

	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	request, err := s.buildTopologyCreateRequest(ctx, instance.HostPort, chal, topology.EntryNodeKey, spec, flag)
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
	imageItem, err := s.imageRepo.FindByID(ctx, chal.ImageID)
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
	ctx context.Context,
	reservedHostPort int,
	chal *model.Challenge,
	entryNodeKey string,
	spec model.TopologySpec,
	flag string,
) (*practiceports.TopologyCreateRequest, error) {
	if len(spec.Nodes) == 0 {
		return nil, errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("challenge topology has no nodes"))
	}
	if chal != nil && chal.InstanceSharing == model.InstanceSharingShared {
		for _, node := range spec.Nodes {
			if node.InjectFlag {
				return nil, errcode.ErrInvalidParams.WithCause(errors.New("共享实例策略不支持带 Flag 注入的拓扑"))
			}
		}
	}

	defaultImageRef, err := s.resolveAvailableImageRef(ctx, chal.ImageID)
	if err != nil {
		return nil, err
	}

	request := &practiceports.TopologyCreateRequest{
		ReservedHostPort: reservedHostPort,
		Networks:         make([]practiceports.TopologyCreateNetwork, 0),
		Nodes:            make([]practiceports.TopologyCreateNode, 0, len(spec.Nodes)),
		Policies:         append([]model.TopologyTrafficPolicy(nil), spec.Policies...),
	}
	runtimePlan := domain.BuildRuntimeTopologyPlan(spec)
	request.Networks = append(request.Networks, runtimePlan.Networks...)
	for _, node := range spec.Nodes {
		imageRef := defaultImageRef
		if node.ImageID > 0 {
			imageRef, err = s.resolveAvailableImageRef(ctx, node.ImageID)
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

		request.Nodes = append(request.Nodes, practiceports.TopologyCreateNode{
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

func (s *Service) resolveAvailableImageRef(ctx context.Context, imageID int64) (string, error) {
	imageItem, err := s.imageRepo.FindByID(ctx, imageID)
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
