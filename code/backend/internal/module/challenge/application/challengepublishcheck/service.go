package challengepublishcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/challenge/application/challengecatalog"
	"ctf-platform/internal/module/challenge/application/challengecore"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type ChallengeSelfChecker interface {
	SelfCheckChallenge(ctx context.Context, id int64) (*challengecontracts.ChallengeSelfCheckResp, error)
}

type ChallengePublisher interface {
	PublishChallenge(ctx context.Context, id int64) error
}

type challengeWriteLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*challengeports.ChallengeWriteModel, error)
}

type Config struct {
	PollInterval time.Duration
	BatchSize    int
}

type ChallengePublishCheckService struct {
	challengeRepo challengeWriteLookupRepository
	jobRepo       challengeports.ChallengePublishCheckRepository
	selfChecker   ChallengeSelfChecker
	publisher     ChallengePublisher
	eventBus      platformevents.Bus
	config        Config
	logger        *zap.Logger
}

func NewChallengePublishCheckService(
	challengeRepo challengeWriteLookupRepository,
	jobRepo challengeports.ChallengePublishCheckRepository,
	selfChecker ChallengeSelfChecker,
	publisher ChallengePublisher,
	cfg Config,
	eventBus platformevents.Bus,
	logger *zap.Logger,
) *ChallengePublishCheckService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1
	}
	return &ChallengePublishCheckService{
		challengeRepo: challengeRepo,
		jobRepo:       jobRepo,
		selfChecker:   selfChecker,
		publisher:     publisher,
		eventBus:      eventBus,
		config:        cfg,
		logger:        logger,
	}
}

func (s *ChallengePublishCheckService) SetEventBus(bus platformevents.Bus) *ChallengePublishCheckService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	if publisher, ok := s.publisher.(interface {
		SetEventBus(platformevents.Bus) *challengecore.ChallengeService
	}); ok {
		publisher.SetEventBus(bus)
	}
	return s
}

func (s *ChallengePublishCheckService) RequestPublishCheck(ctx context.Context, actorUserID, id int64) (*challengecontracts.ChallengePublishCheckJobResp, error) {
	challengeWriteModel, err := s.challengeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, err
	}
	challenge := challengeWriteModel
	if challenge.Status == challengecontracts.ChallengeStatusPublished {
		return nil, apperror.ErrConflict.WithCause(errors.New("题目已发布，无需重复提交发布检查"))
	}

	active, err := s.jobRepo.FindActivePublishCheckJobByChallengeID(ctx, id)
	switch {
	case err == nil:
		return s.buildPublishCheckJobResp(active), nil
	case err != nil && !errors.Is(err, challengeports.ErrChallengePublishCheckJobNotFound):
		return nil, err
	}

	job := &challengeentity.ChallengePublishCheckJob{
		ChallengeID:   challenge.ID,
		RequestedBy:   actorUserID,
		Status:        challengeentity.ChallengePublishCheckStatusPending,
		RequestSource: "admin_publish",
	}
	if err := s.jobRepo.CreatePublishCheckJob(ctx, job); err != nil {
		active, activeErr := s.jobRepo.FindActivePublishCheckJobByChallengeID(ctx, id)
		if activeErr == nil {
			return s.buildPublishCheckJobResp(active), nil
		}
		return nil, err
	}
	return s.buildPublishCheckJobResp(job), nil
}

func (s *ChallengePublishCheckService) GetLatestPublishCheck(ctx context.Context, id int64) (*challengecontracts.ChallengePublishCheckJobResp, error) {
	challengeWriteModel, err := s.challengeRepo.FindByID(ctx, id)
	if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
		return nil, challengecontracts.ErrChallengeNotFound
	}
	if err != nil {
		return nil, err
	}
	challenge := challengeWriteModel
	job, err := s.jobRepo.FindLatestPublishCheckJobByChallengeID(ctx, id)
	if errors.Is(err, challengeports.ErrChallengePublishCheckJobNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if challenge.UpdatedAt.After(job.UpdatedAt) {
		return nil, apperror.ErrNotFound
	}
	return s.buildPublishCheckJobResp(job), nil
}

func (s *ChallengePublishCheckService) RunPublishCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(s.config.PollInterval)
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

func (s *ChallengePublishCheckService) DispatchPublishCheckJobs(ctx context.Context) {
	s.dispatchPublishCheckJobs(ctx)
}

func (s *ChallengePublishCheckService) dispatchPublishCheckJobs(ctx context.Context) {
	jobs, err := s.jobRepo.ListPendingPublishCheckJobs(ctx, s.config.BatchSize)
	if err != nil {
		s.logger.Warn("list pending publish check jobs failed", zap.Error(err))
		return
	}
	for _, job := range jobs {
		if job == nil {
			continue
		}
		startedAt := time.Now().UTC()
		started, err := s.jobRepo.TryStartPublishCheckJob(ctx, job.ID, startedAt)
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

func (s *ChallengePublishCheckService) ProcessPublishCheckJob(ctx context.Context, jobID int64) {
	s.processPublishCheckJob(ctx, jobID)
}

func (s *ChallengePublishCheckService) processPublishCheckJob(ctx context.Context, jobID int64) {
	job, err := s.loadPublishCheckJob(ctx, jobID)
	if err != nil {
		s.logger.Warn("load publish check job failed", zap.Int64("job_id", jobID), zap.Error(err))
		return
	}
	challengeWriteModel, err := s.challengeRepo.FindByID(ctx, job.ChallengeID)
	if err != nil {
		s.finishPublishCheckJob(ctx, job, nil, false, fmt.Sprintf("读取题目失败: %v", err), &challengeports.ChallengeWriteModel{
			ID:    job.ChallengeID,
			Title: fmt.Sprintf("题目 #%d", job.ChallengeID),
		})
		return
	}
	challenge := challengeWriteModel

	resp, err := s.selfChecker.SelfCheckChallenge(ctx, challenge.ID)
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
		if err := s.publisher.PublishChallenge(ctx, challenge.ID); err != nil {
			passed = false
			failureSummary = fmt.Sprintf("自动发布失败: %v", err)
		} else {
			now := time.Now().UTC()
			publishedAt = &now
		}
	}
	s.finishPublishCheckJob(ctx, job, resp, passed, failureSummary, challenge)
	if passed && publishedAt != nil {
		job.PublishedAt = publishedAt
		_ = s.jobRepo.UpdatePublishCheckJob(ctx, job)
	}
}

func (s *ChallengePublishCheckService) loadPublishCheckJob(ctx context.Context, id int64) (*challengeentity.ChallengePublishCheckJob, error) {
	return s.jobRepo.FindPublishCheckJobByID(ctx, id)
}

func (s *ChallengePublishCheckService) finishPublishCheckJob(ctx context.Context, job *challengeentity.ChallengePublishCheckJob, result *challengecontracts.ChallengeSelfCheckResp, passed bool, failureSummary string, challenge *challengeports.ChallengeWriteModel) {
	if job == nil {
		return
	}
	now := time.Now().UTC()
	job.FinishedAt = &now
	job.UpdatedAt = now
	job.FailureSummary = strings.TrimSpace(failureSummary)
	if passed {
		job.Status = challengeentity.ChallengePublishCheckStatusPassed
		job.PublishedAt = &now
	} else {
		job.Status = challengeentity.ChallengePublishCheckStatusFailed
	}
	if result != nil {
		if content, err := json.Marshal(result); err == nil {
			job.ResultJSON = string(content)
		}
	}
	if err := s.jobRepo.UpdatePublishCheckJob(ctx, job); err != nil {
		s.logger.Warn("update publish check job failed", zap.Int64("job_id", job.ID), zap.Error(err))
	}
	if challenge != nil {
		challengecatalog.PublishWeakEvent(ctx, s.logger, s.eventBus, platformevents.Event{
			Name: challengecontracts.EventPublishCheckFinished,
			Payload: challengecontracts.PublishCheckFinishedEvent{
				UserID:         job.RequestedBy,
				ChallengeID:    challenge.ID,
				ChallengeTitle: challenge.Title,
				Passed:         passed,
				FailureSummary: job.FailureSummary,
			},
		})
	}
}

func buildPublishCheckFailureSummary(resp *challengecontracts.ChallengeSelfCheckResp) string {
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

func (s *ChallengePublishCheckService) buildPublishCheckJobResp(job *challengeentity.ChallengePublishCheckJob) *challengecontracts.ChallengePublishCheckJobResp {
	if job == nil {
		return nil
	}
	resp := &challengecontracts.ChallengePublishCheckJobResp{
		ID:             job.ID,
		ChallengeID:    job.ChallengeID,
		RequestedBy:    job.RequestedBy,
		RequestSource:  job.RequestSource,
		FailureSummary: job.FailureSummary,
		StartedAt:      copyTimePtr(job.StartedAt),
		FinishedAt:     copyTimePtr(job.FinishedAt),
		PublishedAt:    copyTimePtr(job.PublishedAt),
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}
	resp.Status = mapPublishCheckStatus(job.Status)
	resp.Active = isActivePublishCheckStatus(job.Status)
	if strings.TrimSpace(job.ResultJSON) != "" {
		var result challengecontracts.ChallengeSelfCheckResp
		if err := json.Unmarshal([]byte(job.ResultJSON), &result); err == nil {
			resp.Result = &result
		}
	}
	return resp
}

func copyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mapPublishCheckStatus(status string) string {
	switch status {
	case challengeentity.ChallengePublishCheckStatusPending:
		return "queued"
	case challengeentity.ChallengePublishCheckStatusPassed:
		return "succeeded"
	default:
		return status
	}
}

func isActivePublishCheckStatus(status string) bool {
	return status == challengeentity.ChallengePublishCheckStatusPending || status == challengeentity.ChallengePublishCheckStatusRunning
}
