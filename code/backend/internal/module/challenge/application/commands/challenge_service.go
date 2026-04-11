package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type SelfCheckConfig struct {
	RuntimeCreateTimeout     time.Duration
	FlagGlobalSecret         string
	PublishCheckPollInterval time.Duration
	PublishCheckBatchSize    int
}

type ChallengeNotificationSender interface {
	SendChallengePublishCheckResult(ctx context.Context, userID int64, challengeID int64, challengeTitle string, passed bool, failureSummary string) error
}

type ChallengeService struct {
	db            *gorm.DB
	repo          challengeports.ChallengeCommandRepository
	imageRepo     challengeports.ImageRepository
	topologyRepo  challengeports.ChallengeTopologyRepository
	runtimeProbe  challengeports.ChallengeRuntimeProbe
	notifications ChallengeNotificationSender
	selfCheckCfg  SelfCheckConfig
	logger        *zap.Logger
}

func NewChallengeService(
	db *gorm.DB,
	repo challengeports.ChallengeCommandRepository,
	imageRepo challengeports.ImageRepository,
	topologyRepo challengeports.ChallengeTopologyRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
	cfg SelfCheckConfig,
	logger *zap.Logger,
	notifications ...ChallengeNotificationSender,
) *ChallengeService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg.RuntimeCreateTimeout <= 0 {
		cfg.RuntimeCreateTimeout = 60 * time.Second
	}
	if cfg.PublishCheckPollInterval <= 0 {
		cfg.PublishCheckPollInterval = 2 * time.Second
	}
	if cfg.PublishCheckBatchSize <= 0 {
		cfg.PublishCheckBatchSize = 1
	}
	return &ChallengeService{
		db:            db,
		repo:          repo,
		imageRepo:     imageRepo,
		topologyRepo:  topologyRepo,
		runtimeProbe:  runtimeProbe,
		notifications: firstChallengeNotificationSender(notifications),
		selfCheckCfg:  cfg,
		logger:        logger,
	}
}

func firstChallengeNotificationSender(senders []ChallengeNotificationSender) ChallengeNotificationSender {
	if len(senders) == 0 {
		return nil
	}
	return senders[0]
}

func (s *ChallengeService) CreateChallenge(actorUserID int64, req *dto.CreateChallengeReq) (*dto.ChallengeResp, error) {
	if req.ImageID > 0 {
		if _, err := s.imageRepo.FindByID(req.ImageID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
			}
			return nil, err
		}
	}

	challenge := &model.Challenge{
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Difficulty:      req.Difficulty,
		Points:          req.Points,
		ImageID:         req.ImageID,
		AttachmentURL:   strings.TrimSpace(req.AttachmentURL),
		InstanceSharing: normalizeInstanceSharing(req.InstanceSharing),
		Status:          model.ChallengeStatusDraft,
		CreatedBy:       &actorUserID,
	}

	hints, err := domain.NormalizeHintModels(req.Hints)
	if err != nil {
		return nil, err
	}
	if err := s.validateInstanceSharingConfig(challenge); err != nil {
		return nil, err
	}
	if err := s.repo.CreateWithHints(challenge, hints); err != nil {
		return nil, err
	}
	return domain.ChallengeRespFromModel(challenge, hints), nil
}

func (s *ChallengeService) UpdateChallenge(id int64, req *dto.UpdateChallengeReq) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	if req.Title != "" {
		challenge.Title = req.Title
	}
	if req.Description != "" {
		challenge.Description = req.Description
	}
	if req.Category != "" {
		challenge.Category = req.Category
	}
	if req.Difficulty != "" {
		challenge.Difficulty = req.Difficulty
	}
	if req.Points > 0 {
		challenge.Points = req.Points
	}
	if req.ImageID != nil {
		if *req.ImageID > 0 {
			if _, err := s.imageRepo.FindByID(*req.ImageID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errcode.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
				}
				return err
			}
		}
		challenge.ImageID = *req.ImageID
	}
	if req.AttachmentURL != nil {
		challenge.AttachmentURL = strings.TrimSpace(*req.AttachmentURL)
	}
	if req.InstanceSharing != "" {
		challenge.InstanceSharing = normalizeInstanceSharing(req.InstanceSharing)
	}

	replaceHints := req.Hints != nil
	hints, err := domain.NormalizeHintModels(req.Hints)
	if err != nil {
		return err
	}
	if err := s.validateInstanceSharingConfig(challenge); err != nil {
		return err
	}

	return s.repo.UpdateWithHints(challenge, hints, replaceHints)
}

func normalizeInstanceSharing(value string) string {
	switch strings.TrimSpace(value) {
	case model.InstanceSharingPerTeam:
		return model.InstanceSharingPerTeam
	case model.InstanceSharingShared:
		return model.InstanceSharingShared
	default:
		return model.InstanceSharingPerUser
	}
}

func (s *ChallengeService) validateInstanceSharingConfig(challenge *model.Challenge) error {
	if challenge == nil || challenge.InstanceSharing != model.InstanceSharingShared {
		return nil
	}
	if challenge.FlagType == model.FlagTypeDynamic {
		return errcode.ErrInvalidParams.WithCause(errors.New("共享实例只适用于无状态题，不支持动态 Flag"))
	}
	if s.topologyRepo == nil || challenge.ID <= 0 {
		return nil
	}
	topology, err := s.topologyRepo.FindChallengeTopologyByChallengeID(challenge.ID)
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	default:
		return err
	}

	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		return errcode.ErrInvalidParams.WithCause(err)
	}
	for _, node := range spec.Nodes {
		if node.InjectFlag {
			return errcode.ErrInvalidParams.WithCause(errors.New("共享实例只适用于无状态题，不支持带 Flag 注入的拓扑"))
		}
	}
	return nil
}

func (s *ChallengeService) DeleteChallenge(id int64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	hasInstances, err := s.repo.HasRunningInstances(id)
	if err != nil {
		return err
	}
	if hasInstances {
		return errcode.New(errcode.ErrConflict.Code, domain.ErrMsgHasRunningStudents, errcode.ErrConflict.HTTPStatus).
			WithCause(errors.New(domain.ErrMsgHasRunningInstances))
	}
	return s.repo.Delete(id)
}

func (s *ChallengeService) PublishChallenge(id int64) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	challenge.Status = model.ChallengeStatusPublished
	return s.repo.Update(challenge)
}

func (s *ChallengeService) RequestPublishCheck(ctx context.Context, actorUserID, id int64) (*dto.ChallengePublishCheckJobResp, error) {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if challenge.Status == model.ChallengeStatusPublished {
		return nil, errcode.ErrConflict.WithCause(errors.New("题目已发布，无需重复提交发布检查"))
	}
	if ctx == nil {
		ctx = context.Background()
	}

	active, err := s.repo.FindActivePublishCheckJobByChallengeID(ctx, id)
	switch {
	case err == nil:
		return s.buildPublishCheckJobResp(active), nil
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	job := &model.ChallengePublishCheckJob{
		ChallengeID:   challenge.ID,
		RequestedBy:   actorUserID,
		Status:        model.ChallengePublishCheckStatusPending,
		RequestSource: "admin_publish",
	}
	if err := s.repo.CreatePublishCheckJob(ctx, job); err != nil {
		active, activeErr := s.repo.FindActivePublishCheckJobByChallengeID(ctx, id)
		if activeErr == nil {
			return s.buildPublishCheckJobResp(active), nil
		}
		return nil, err
	}
	return s.buildPublishCheckJobResp(job), nil
}

func (s *ChallengeService) GetLatestPublishCheck(ctx context.Context, id int64) (*dto.ChallengePublishCheckJobResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	job, err := s.repo.FindLatestPublishCheckJobByChallengeID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.buildPublishCheckJobResp(job), nil
}

func (s *ChallengeService) RunPublishCheckLoop(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(s.selfCheckCfg.PublishCheckPollInterval)
	defer ticker.Stop()

	for {
		s.dispatchPublishCheckJobs(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *ChallengeService) dispatchPublishCheckJobs(ctx context.Context) {
	jobs, err := s.repo.ListPendingPublishCheckJobs(ctx, s.selfCheckCfg.PublishCheckBatchSize)
	if err != nil {
		s.logger.Warn("list pending publish check jobs failed", zap.Error(err))
		return
	}
	for _, job := range jobs {
		if job == nil {
			continue
		}
		startedAt := time.Now()
		started, err := s.repo.TryStartPublishCheckJob(ctx, job.ID, startedAt)
		if err != nil {
			s.logger.Warn("start publish check job failed", zap.Int64("job_id", job.ID), zap.Error(err))
			continue
		}
		if !started {
			continue
		}
		s.processPublishCheckJob(ctx, job.ID)
	}
}

func (s *ChallengeService) processPublishCheckJob(ctx context.Context, jobID int64) {
	job, err := s.loadPublishCheckJob(ctx, jobID)
	if err != nil {
		s.logger.Warn("load publish check job failed", zap.Int64("job_id", jobID), zap.Error(err))
		return
	}
	challenge, err := s.repo.FindByID(job.ChallengeID)
	if err != nil {
		s.finishPublishCheckJob(ctx, job, nil, false, fmt.Sprintf("读取题目失败: %v", err), &model.Challenge{
			ID:    job.ChallengeID,
			Title: fmt.Sprintf("题目 #%d", job.ChallengeID),
		})
		return
	}

	resp, err := s.SelfCheckChallenge(ctx, challenge.ID)
	if err != nil {
		s.finishPublishCheckJob(ctx, job, nil, false, fmt.Sprintf("执行自检失败: %v", err), challenge)
		return
	}

	passed := resp.Precheck.Passed && resp.Runtime.Passed
	failureSummary := ""
	if !passed {
		failureSummary = buildPublishCheckFailureSummary(resp)
	}

	var publishedAt *time.Time
	if passed {
		if err := s.PublishChallenge(challenge.ID); err != nil {
			passed = false
			failureSummary = fmt.Sprintf("自动发布失败: %v", err)
		} else {
			now := time.Now()
			publishedAt = &now
		}
	}
	s.finishPublishCheckJob(ctx, job, resp, passed, failureSummary, challenge)
	if passed && publishedAt != nil {
		job.PublishedAt = publishedAt
		_ = s.repo.UpdatePublishCheckJob(ctx, job)
	}
}

func (s *ChallengeService) loadPublishCheckJob(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error) {
	return s.repo.FindPublishCheckJobByID(ctx, id)
}

func (s *ChallengeService) finishPublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob, result *dto.ChallengeSelfCheckResp, passed bool, failureSummary string, challenge *model.Challenge) {
	if job == nil {
		return
	}
	now := time.Now()
	job.FinishedAt = &now
	job.UpdatedAt = now
	job.FailureSummary = strings.TrimSpace(failureSummary)
	if passed {
		job.Status = model.ChallengePublishCheckStatusPassed
		job.PublishedAt = &now
	} else {
		job.Status = model.ChallengePublishCheckStatusFailed
	}
	if result != nil {
		if content, err := json.Marshal(result); err == nil {
			job.ResultJSON = string(content)
		}
	}
	if err := s.repo.UpdatePublishCheckJob(ctx, job); err != nil {
		s.logger.Warn("update publish check job failed", zap.Int64("job_id", job.ID), zap.Error(err))
	}
	if s.notifications != nil && challenge != nil {
		if err := s.notifications.SendChallengePublishCheckResult(ctx, job.RequestedBy, challenge.ID, challenge.Title, passed, job.FailureSummary); err != nil {
			s.logger.Warn("send publish check notification failed", zap.Int64("job_id", job.ID), zap.Error(err))
		}
	}
}

func buildPublishCheckFailureSummary(resp *dto.ChallengeSelfCheckResp) string {
	if resp == nil {
		return "平台自检失败"
	}
	for _, step := range resp.Precheck.Steps {
		if !step.Passed && strings.TrimSpace(step.Message) != "" {
			return step.Message
		}
	}
	for _, step := range resp.Runtime.Steps {
		if !step.Passed && strings.TrimSpace(step.Message) != "" {
			return step.Message
		}
	}
	if !resp.Precheck.Passed {
		return "预检未通过"
	}
	if !resp.Runtime.Passed {
		return "运行时自检未通过"
	}
	return ""
}

func (s *ChallengeService) buildPublishCheckJobResp(job *model.ChallengePublishCheckJob) *dto.ChallengePublishCheckJobResp {
	if job == nil {
		return nil
	}
	resp := &dto.ChallengePublishCheckJobResp{
		ID:             job.ID,
		ChallengeID:    job.ChallengeID,
		RequestedBy:    job.RequestedBy,
		Status:         mapPublishCheckStatus(job.Status),
		Active:         isActivePublishCheckStatus(job.Status),
		RequestSource:  job.RequestSource,
		FailureSummary: job.FailureSummary,
		StartedAt:      job.StartedAt,
		FinishedAt:     job.FinishedAt,
		PublishedAt:    job.PublishedAt,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}
	if strings.TrimSpace(job.ResultJSON) != "" {
		var result dto.ChallengeSelfCheckResp
		if err := json.Unmarshal([]byte(job.ResultJSON), &result); err == nil {
			resp.Result = &result
		}
	}
	return resp
}

func mapPublishCheckStatus(status string) string {
	switch status {
	case model.ChallengePublishCheckStatusPending:
		return "queued"
	case model.ChallengePublishCheckStatusPassed:
		return "succeeded"
	default:
		return status
	}
}

func isActivePublishCheckStatus(status string) bool {
	return status == model.ChallengePublishCheckStatusPending || status == model.ChallengePublishCheckStatusRunning
}

type challengeSelfCheckRuntimeInput struct {
	defaultImageRef string
	nodeImageRefs   map[int64]string
	entryNodeKey    string
	topologySpec    model.TopologySpec
	useTopology     bool
}

func (s *ChallengeService) SelfCheckChallenge(ctx context.Context, id int64) (*dto.ChallengeSelfCheckResp, error) {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	resp := &dto.ChallengeSelfCheckResp{
		ChallengeID: challenge.ID,
	}

	resp.Precheck.StartedAt = time.Now()
	input, precheckPassed, err := s.runPrecheck(challenge, &resp.Precheck.Steps)
	resp.Precheck.EndedAt = time.Now()
	if err != nil {
		return nil, err
	}
	resp.Precheck.Passed = precheckPassed

	resp.Runtime.StartedAt = time.Now()
	if !resp.Precheck.Passed {
		resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
			Name:    "runtime_startup",
			Passed:  false,
			Message: "预检未通过，已跳过真实拉起",
		})
		resp.Runtime.EndedAt = time.Now()
		return resp, nil
	}
	if s.runtimeProbe == nil {
		resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
			Name:    "runtime_startup",
			Passed:  false,
			Message: "运行时自测能力未配置",
		})
		resp.Runtime.EndedAt = time.Now()
		return resp, nil
	}

	if ctx == nil {
		ctx = context.Background()
	}
	createCtx, cancel := context.WithTimeout(ctx, s.selfCheckCfg.RuntimeCreateTimeout)
	defer cancel()

	runtimePassed := true
	flag, flagErr := s.buildRuntimeFlag(challenge)
	if flagErr != nil {
		runtimePassed = false
		resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
			Name:    "runtime_flag",
			Passed:  false,
			Message: fmt.Sprintf("生成运行时 Flag 失败: %v", flagErr),
		})
	} else {
		resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
			Name:    "runtime_flag",
			Passed:  true,
			Message: "运行时 Flag 已准备",
		})
	}

	var (
		runtimeDetails model.InstanceRuntimeDetails
		accessURL      string
	)
	if runtimePassed {
		if input.useTopology {
			req, buildErr := s.buildTopologyRuntimeRequest(input, flag)
			if buildErr != nil {
				runtimePassed = false
				resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
					Name:    "runtime_startup",
					Passed:  false,
					Message: fmt.Sprintf("构建拓扑启动请求失败: %v", buildErr),
				})
			} else {
				result, startupErr := s.runtimeProbe.CreateTopology(createCtx, req)
				if startupErr != nil {
					runtimePassed = false
					resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
						Name:    "runtime_startup",
						Passed:  false,
						Message: fmt.Sprintf("拓扑拉起失败: %v", startupErr),
					})
				} else {
					accessURL = result.AccessURL
					runtimeDetails = result.RuntimeDetails
					resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
						Name:    "runtime_startup",
						Passed:  true,
						Message: "拓扑实例拉起成功",
					})
				}
			}
		} else {
			startupAccessURL, details, startupErr := s.runtimeProbe.CreateContainer(createCtx, input.defaultImageRef, map[string]string{
				"FLAG": flag,
			})
			if startupErr != nil {
				runtimePassed = false
				resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
					Name:    "runtime_startup",
					Passed:  false,
					Message: fmt.Sprintf("单容器拉起失败: %v", startupErr),
				})
			} else {
				accessURL = startupAccessURL
				runtimeDetails = details
				resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
					Name:    "runtime_startup",
					Passed:  true,
					Message: "单容器实例拉起成功",
				})
			}
		}
	}

	if runtimePassed {
		if cleanupErr := s.runtimeProbe.CleanupRuntimeDetails(ctx, runtimeDetails); cleanupErr != nil {
			runtimePassed = false
			resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
				Name:    "runtime_cleanup",
				Passed:  false,
				Message: fmt.Sprintf("运行时资源清理失败: %v", cleanupErr),
			})
		} else {
			resp.Runtime.Steps = append(resp.Runtime.Steps, dto.ChallengeSelfCheckStepResp{
				Name:    "runtime_cleanup",
				Passed:  true,
				Message: "运行时资源已清理",
			})
		}
	}

	resp.Runtime.EndedAt = time.Now()
	resp.Runtime.Passed = runtimePassed
	resp.Runtime.AccessURL = accessURL
	resp.Runtime.ContainerCount = len(runtimeDetails.Containers)
	resp.Runtime.NetworkCount = len(runtimeDetails.Networks)
	return resp, nil
}

func (s *ChallengeService) runPrecheck(challenge *model.Challenge, steps *[]dto.ChallengeSelfCheckStepResp) (challengeSelfCheckRuntimeInput, bool, error) {
	input := challengeSelfCheckRuntimeInput{
		nodeImageRefs: make(map[int64]string),
	}
	passed := true

	flagOK, flagMessage := s.validateFlagConfig(challenge)
	*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
		Name:    "flag_config",
		Passed:  flagOK,
		Message: flagMessage,
	})
	if !flagOK {
		passed = false
	}

	if challenge.ImageID > 0 {
		imageRef, err := s.resolveAvailableImageRef(challenge.ImageID)
		if err != nil {
			passed = false
			*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
				Name:    "challenge_image",
				Passed:  false,
				Message: fmt.Sprintf("题目默认镜像不可用: %v", err),
			})
		} else {
			input.defaultImageRef = imageRef
			*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
				Name:    "challenge_image",
				Passed:  true,
				Message: "题目默认镜像可用",
			})
		}
	} else {
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "challenge_image",
			Passed:  true,
			Message: "题目未配置默认镜像",
		})
	}

	if s.topologyRepo == nil {
		return input, false, errcode.ErrInternal.WithCause(errors.New("challenge topology repository is not configured"))
	}
	topology, err := s.topologyRepo.FindChallengeTopologyByChallengeID(challenge.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return input, false, err
		}
		if challenge.ImageID <= 0 {
			passed = false
			*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
				Name:    "topology_or_single_container",
				Passed:  false,
				Message: "未配置拓扑且题目默认镜像为空，无法执行真实拉起",
			})
		} else {
			*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
				Name:    "topology_or_single_container",
				Passed:  true,
				Message: "未配置拓扑，将按单容器模式自测",
			})
		}
		return input, passed, nil
	}

	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_spec",
			Passed:  false,
			Message: fmt.Sprintf("拓扑解码失败: %v", err),
		})
		return input, false, nil
	}
	if len(spec.Nodes) == 0 {
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_spec",
			Passed:  false,
			Message: "拓扑至少需要一个节点",
		})
		return input, false, nil
	}

	entryKey := strings.TrimSpace(topology.EntryNodeKey)
	if entryKey == "" {
		entryKey = spec.Nodes[0].Key
	}
	entryPortOK := false
	needsDefaultImage := false
	for _, node := range spec.Nodes {
		if node.Key == entryKey && node.ServicePort > 0 {
			entryPortOK = true
		}
		if node.ImageID == 0 {
			needsDefaultImage = true
			continue
		}
		if _, exists := input.nodeImageRefs[node.ImageID]; exists {
			continue
		}
		nodeImageRef, resolveErr := s.resolveAvailableImageRef(node.ImageID)
		if resolveErr != nil {
			passed = false
			*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
				Name:    "topology_images",
				Passed:  false,
				Message: fmt.Sprintf("拓扑节点镜像不可用 (image_id=%d): %v", node.ImageID, resolveErr),
			})
			break
		}
		input.nodeImageRefs[node.ImageID] = nodeImageRef
	}
	if passed {
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_images",
			Passed:  true,
			Message: "拓扑节点镜像检查通过",
		})
	}

	if needsDefaultImage && input.defaultImageRef == "" {
		passed = false
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_default_image",
			Passed:  false,
			Message: "拓扑存在未指定 image_id 的节点，但题目默认镜像不可用",
		})
	} else {
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_default_image",
			Passed:  true,
			Message: "拓扑默认镜像策略检查通过",
		})
	}

	if !entryPortOK {
		passed = false
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_entry",
			Passed:  false,
			Message: "拓扑入口节点不存在或未设置 service_port",
		})
	} else {
		*steps = append(*steps, dto.ChallengeSelfCheckStepResp{
			Name:    "topology_entry",
			Passed:  true,
			Message: "拓扑入口节点配置有效",
		})
	}

	input.useTopology = true
	input.topologySpec = spec
	input.entryNodeKey = entryKey
	return input, passed, nil
}

func (s *ChallengeService) validateFlagConfig(challenge *model.Challenge) (bool, string) {
	switch challenge.FlagType {
	case model.FlagTypeStatic:
		if challenge.FlagHash == "" || challenge.FlagSalt == "" {
			return false, "静态 Flag 未正确配置（缺少 hash/salt）"
		}
		return true, "静态 Flag 配置有效"
	case model.FlagTypeDynamic:
		if strings.TrimSpace(s.selfCheckCfg.FlagGlobalSecret) == "" {
			return false, "动态 Flag 依赖的全局密钥未配置"
		}
		return true, "动态 Flag 配置有效"
	case model.FlagTypeRegex:
		if _, err := regexp.Compile(strings.TrimSpace(challenge.FlagRegex)); err != nil {
			return false, fmt.Sprintf("Regex Flag 配置无效: %v", err)
		}
		return true, "Regex Flag 配置有效"
	case model.FlagTypeManualReview:
		return true, "人工审核题已跳过 Flag 自动校验"
	default:
		return false, "Flag 类型无效"
	}
}

func (s *ChallengeService) buildRuntimeFlag(challenge *model.Challenge) (string, error) {
	switch challenge.FlagType {
	case model.FlagTypeStatic:
		return challenge.FlagHash, nil
	case model.FlagTypeDynamic:
		nonce, err := crypto.GenerateNonce()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(s.selfCheckCfg.FlagGlobalSecret) == "" {
			return "", fmt.Errorf("flag global secret is empty")
		}
		return crypto.GenerateDynamicFlag(0, challenge.ID, s.selfCheckCfg.FlagGlobalSecret, nonce, challenge.FlagPrefix), nil
	case model.FlagTypeRegex, model.FlagTypeManualReview:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported flag type %s", challenge.FlagType)
	}
}

func (s *ChallengeService) buildTopologyRuntimeRequest(
	input challengeSelfCheckRuntimeInput,
	flag string,
) (*challengeports.RuntimeTopologyCreateRequest, error) {
	req := &challengeports.RuntimeTopologyCreateRequest{
		Networks: make([]challengeports.RuntimeTopologyCreateNetwork, 0, len(input.topologySpec.Networks)),
		Nodes:    make([]challengeports.RuntimeTopologyCreateNode, 0, len(input.topologySpec.Nodes)),
		Policies: append([]model.TopologyTrafficPolicy(nil), input.topologySpec.Policies...),
	}
	for _, network := range input.topologySpec.Networks {
		req.Networks = append(req.Networks, challengeports.RuntimeTopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	defaultNetworkKey := model.TopologyDefaultNetworkKey
	if len(req.Networks) > 0 {
		defaultNetworkKey = req.Networks[0].Key
	}

	for _, node := range input.topologySpec.Nodes {
		imageRef := input.defaultImageRef
		if node.ImageID > 0 {
			imageRef = input.nodeImageRefs[node.ImageID]
		}
		if imageRef == "" {
			return nil, fmt.Errorf("node %s image is empty", node.Key)
		}

		env := make(map[string]string, len(node.Env)+1)
		for key, value := range node.Env {
			env[key] = value
		}
		if node.InjectFlag {
			env["FLAG"] = flag
		}

		networkKeys := append([]string(nil), node.NetworkKeys...)
		if len(networkKeys) == 0 {
			networkKeys = []string{defaultNetworkKey}
		}

		var resources *model.ResourceLimits
		if node.Resources != nil {
			resources = &model.ResourceLimits{
				CPUQuota:  node.Resources.CPUQuota,
				Memory:    node.Resources.MemoryMB * 1024 * 1024,
				PidsLimit: node.Resources.PidsLimit,
			}
		}

		req.Nodes = append(req.Nodes, challengeports.RuntimeTopologyCreateNode{
			Key:          node.Key,
			Image:        imageRef,
			Env:          env,
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.Key == input.entryNodeKey,
			NetworkKeys:  networkKeys,
			Resources:    resources,
		})
	}
	if len(req.Networks) == 0 {
		req.Networks = []challengeports.RuntimeTopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		}
	}
	return req, nil
}

func (s *ChallengeService) resolveAvailableImageRef(imageID int64) (string, error) {
	if imageID <= 0 {
		return "", fmt.Errorf("invalid image id")
	}
	imageItem, err := s.imageRepo.FindByID(imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errcode.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
		}
		return "", err
	}
	if imageItem.Status != model.ImageStatusAvailable {
		return "", fmt.Errorf("image %d status=%s", imageItem.ID, imageItem.Status)
	}
	return fmt.Sprintf("%s:%s", imageItem.Name, imageItem.Tag), nil
}
