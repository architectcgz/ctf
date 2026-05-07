package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

const awdCheckerPreviewAttemptCount = 3

func (s *AWDService) RunCurrentRoundChecks(ctx context.Context, contestID int64, req RunCurrentRoundChecksInput) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if !now.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}
	if err := ensureAWDReadinessGate(ctx, s.repo, contestID, req.ForceOverride, req.OverrideReason); err != nil {
		return nil, err
	}
	round, err := s.resolveCurrentRoundForContest(ctx, contest)
	if err != nil {
		return nil, err
	}

	if s.roundManager == nil {
		return nil, errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.RunRoundServiceChecks(ctx, contest, round, contestdomain.AWDCheckSourceManualCurrent); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return s.buildCheckerRunResp(ctx, contestID, round)
}

func (s *AWDService) RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}

	if s.roundManager == nil {
		return nil, errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.RunRoundServiceChecks(ctx, contest, round, contestdomain.AWDCheckSourceManualSelected); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return s.buildCheckerRunResp(ctx, contestID, round)
}

func (s *AWDService) PreviewChecker(ctx context.Context, contestID int64, req PreviewCheckerInput) (*dto.AWDCheckerPreviewResp, error) {
	s.reportAWDPreviewProgress(ctx, contestID, req.PreviewRequestID, "prepare", "准备预览环境", "正在校验当前 Checker 草稿，并准备目标访问上下文。", 0, awdCheckerPreviewAttemptCount, "running", nil)

	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "prepare", "读取赛事配置失败", err)
		return nil, err
	}

	var previewServiceID int64
	previewChallengeID := req.AWDChallengeID
	var previewService *model.ContestAWDService
	if req.ServiceID > 0 {
		service, err := s.resolveContestRuntimeService(ctx, contestID, req.ServiceID)
		if err != nil {
			return nil, err
		}
		previewService = service
		previewServiceID = service.ID
		previewChallengeID = service.AWDChallengeID
		if req.AWDChallengeID > 0 && req.AWDChallengeID != service.AWDChallengeID {
			return nil, errcode.ErrInvalidParams
		}
	}
	if previewChallengeID <= 0 {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "prepare", "缺少试跑目标题目。", errcode.ErrInvalidParams)
		return nil, errcode.ErrInvalidParams
	}

	checkerType, checkerConfig, err := validateAndNormalizeContestAWDFields(
		contest,
		req.CheckerType,
		req.CheckerConfig,
		0,
		0,
	)
	if err != nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "prepare", "Checker 配置无效，无法开始试跑。", err)
		return nil, err
	}
	if checkerType == "" {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "prepare", "Checker 类型无效。", errcode.ErrInvalidParams)
		return nil, errcode.ErrInvalidParams
	}

	previewAccessURL, checkerTokenEnv, checkerToken, cleanupRuntime, err := s.prepareCheckerPreviewAccessURL(
		ctx,
		contestID,
		previewService,
		previewChallengeID,
		req.AccessURL,
		req.PreviewFlag,
	)
	if err != nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "prepare", "预览环境准备失败。", err)
		return nil, err
	}

	if s.roundManager == nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "prepare", "试跑执行器不可用。", errors.New("awd round manager is nil"))
		return nil, errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}

	preview, err := s.runPreviewCheckerAttempts(ctx, contestID, req.PreviewRequestID, contestports.AWDServicePreviewRequest{
		ServiceID:       previewServiceID,
		AWDChallengeID:  previewChallengeID,
		CheckerType:     checkerType,
		CheckerConfig:   checkerConfig,
		CheckerTokenEnv: checkerTokenEnv,
		CheckerToken:    checkerToken,
		AccessURL:       previewAccessURL,
		PreviewFlag:     req.PreviewFlag,
	})
	if cleanupErr := s.cleanupCheckerPreviewRuntime(ctx, cleanupRuntime, err); cleanupErr != nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "summary", "预览实例回收失败。", cleanupErr)
		return nil, cleanupErr
	}
	if err != nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "summary", "试跑执行失败。", err)
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.AWDCheckerPreviewResp{
		CheckerType:   preview.CheckerType,
		ServiceStatus: preview.ServiceStatus,
		CheckResult:   contestdomain.ParseAWDCheckResult(preview.CheckResult),
		PreviewContext: dto.AWDCheckerPreviewContextResp{
			ServiceID:      preview.PreviewContext.ServiceID,
			AccessURL:      preview.PreviewContext.AccessURL,
			PreviewFlag:    preview.PreviewContext.PreviewFlag,
			RoundNumber:    preview.PreviewContext.RoundNumber,
			TeamID:         preview.PreviewContext.TeamID,
			AWDChallengeID: preview.PreviewContext.AWDChallengeID,
		},
	}
	if s.redis == nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "summary", "试跑结果暂不可保存。", errcode.ErrAWDCheckerPreviewUnavailable)
		return nil, errcode.ErrAWDCheckerPreviewUnavailable
	}
	previewToken, err := storeAWDCheckerPreviewToken(ctx, s.redis, contestID, previewServiceID, previewChallengeID, checkerType, checkerConfig, checkerTokenEnv, resp)
	if err != nil {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "summary", "试跑结果保存失败。", err)
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if strings.TrimSpace(previewToken) == "" {
		s.reportAWDPreviewFailure(ctx, contestID, req.PreviewRequestID, "summary", "试跑结果未生成可保存 token。", errcode.ErrAWDCheckerPreviewUnavailable)
		return nil, errcode.ErrAWDCheckerPreviewUnavailable
	}
	resp.PreviewToken = previewToken
	return resp, nil
}

func (s *AWDService) runPreviewCheckerAttempts(ctx context.Context, contestID int64, requestID string, req contestports.AWDServicePreviewRequest) (*contestports.AWDServicePreviewResult, error) {
	results := make([]*contestports.AWDServicePreviewResult, 0, awdCheckerPreviewAttemptCount)
	for i := 0; i < awdCheckerPreviewAttemptCount; i++ {
		attempt := i + 1
		s.reportAWDPreviewProgress(
			ctx,
			contestID,
			requestID,
			fmt.Sprintf("attempt-%d", attempt),
			fmt.Sprintf("第 %d 轮试跑", attempt),
			fmt.Sprintf("正在执行第 %d / %d 轮请求校验。", attempt, awdCheckerPreviewAttemptCount),
			attempt,
			awdCheckerPreviewAttemptCount,
			"running",
			nil,
		)
		preview, err := s.roundManager.PreviewServiceCheck(ctx, req)
		if err != nil {
			s.reportAWDPreviewFailure(
				ctx,
				contestID,
				requestID,
				fmt.Sprintf("attempt-%d", attempt),
				fmt.Sprintf("第 %d 轮试跑失败。", attempt),
				err,
			)
			return nil, err
		}
		results = append(results, preview)
	}
	s.reportAWDPreviewProgress(ctx, contestID, requestID, "summary", "汇总结果", "正在整理三轮试跑结果并生成摘要。", 0, awdCheckerPreviewAttemptCount, "running", nil)
	return aggregateAWDCheckerPreviewResults(results)
}

func (s *AWDService) reportAWDPreviewProgress(
	ctx context.Context,
	contestID int64,
	requestID string,
	phaseKey string,
	phaseLabel string,
	detail string,
	attempt int,
	totalAttempts int,
	status string,
	extra map[string]any,
) {
	broadcastAWDPreviewProgress(ctx, s.broadcaster, contestID, requestID, phaseKey, phaseLabel, detail, attempt, totalAttempts, status, extra)
}

func (s *AWDService) reportAWDPreviewFailure(ctx context.Context, contestID int64, requestID string, phaseKey string, detail string, err error) {
	if err == nil {
		return
	}
	s.reportAWDPreviewProgress(ctx, contestID, requestID, phaseKey, "试跑失败", detail, 0, awdCheckerPreviewAttemptCount, "failed", map[string]any{
		"error": err.Error(),
	})
}

func aggregateAWDCheckerPreviewResults(results []*contestports.AWDServicePreviewResult) (*contestports.AWDServicePreviewResult, error) {
	if len(results) == 0 {
		return nil, errors.New("preview checker returned no results")
	}

	totalCount := len(results)
	passCount := 0
	for _, item := range results {
		if item != nil && item.ServiceStatus == model.AWDServiceStatusUp {
			passCount++
		}
	}

	requiredCount := totalCount/2 + 1
	quorumPassed := passCount >= requiredCount
	representative := selectAWDCheckerPreviewRepresentative(results, quorumPassed)
	if representative == nil {
		return nil, errors.New("preview checker returned nil result")
	}

	result := cloneAWDCheckerPreviewMap(contestdomain.ParseAWDCheckResult(representative.CheckResult))
	attempts := make([]map[string]any, 0, len(results))
	for index, item := range results {
		attemptResult := cloneAWDCheckerPreviewMap(contestdomain.ParseAWDCheckResult(item.CheckResult))
		attemptResult["attempt"] = index + 1
		attemptResult["service_status"] = item.ServiceStatus
		attempts = append(attempts, attemptResult)
	}

	result["check_source"] = "checker_preview"
	result["preview_pass_count"] = passCount
	result["preview_total_count"] = totalCount
	result["preview_required_count"] = requiredCount
	result["preview_summary"] = fmt.Sprintf("%d/%d 通过", passCount, totalCount)
	result["preview_attempts"] = attempts
	if checkedAt := latestAWDCheckerPreviewCheckedAt(results); checkedAt != "" {
		result["checked_at"] = checkedAt
	}
	if quorumPassed {
		result["status_reason"] = "preview_quorum_passed"
	} else {
		result["status_reason"] = "preview_quorum_failed"
	}

	checkResult, err := contestdomain.MarshalAWDCheckResult(result)
	if err != nil {
		return nil, err
	}

	serviceStatus := model.AWDServiceStatusDown
	if quorumPassed {
		serviceStatus = model.AWDServiceStatusUp
	}

	return &contestports.AWDServicePreviewResult{
		ServiceStatus:  serviceStatus,
		CheckerType:    representative.CheckerType,
		CheckResult:    checkResult,
		PreviewContext: representative.PreviewContext,
	}, nil
}

func selectAWDCheckerPreviewRepresentative(results []*contestports.AWDServicePreviewResult, quorumPassed bool) *contestports.AWDServicePreviewResult {
	if quorumPassed {
		for _, item := range results {
			if item != nil && item.ServiceStatus == model.AWDServiceStatusUp {
				return item
			}
		}
	}
	for _, item := range results {
		if item != nil && item.ServiceStatus != model.AWDServiceStatusUp {
			return item
		}
	}
	for _, item := range results {
		if item != nil {
			return item
		}
	}
	return nil
}

func cloneAWDCheckerPreviewMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func latestAWDCheckerPreviewCheckedAt(results []*contestports.AWDServicePreviewResult) string {
	for i := len(results) - 1; i >= 0; i-- {
		item := results[i]
		if item == nil {
			continue
		}
		checkedAt, _ := contestdomain.ParseAWDCheckResult(item.CheckResult)["checked_at"].(string)
		if strings.TrimSpace(checkedAt) != "" {
			return checkedAt
		}
	}
	return ""
}
