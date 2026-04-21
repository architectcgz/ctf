package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) RunCurrentRoundChecks(ctx context.Context, contestID int64, req *dto.RunCurrentAWDCheckerReq) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}
	if req == nil {
		req = &dto.RunCurrentAWDCheckerReq{}
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

func (s *AWDService) PreviewChecker(ctx context.Context, contestID int64, req *dto.PreviewAWDCheckerReq) (*dto.AWDCheckerPreviewResp, error) {
	if req == nil {
		return nil, errcode.ErrInvalidParams
	}

	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}

	var previewServiceID int64
	previewChallengeID := req.ChallengeID
	var previewService *model.ContestAWDService
	if req.ServiceID > 0 {
		service, err := s.resolveContestRuntimeService(ctx, contestID, req.ServiceID)
		if err != nil {
			return nil, err
		}
		previewService = service
		previewServiceID = service.ID
		previewChallengeID = service.ChallengeID
		if req.ChallengeID > 0 && req.ChallengeID != service.ChallengeID {
			return nil, errcode.ErrInvalidParams
		}
	}
	if previewChallengeID <= 0 {
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
		return nil, err
	}
	if checkerType == "" {
		return nil, errcode.ErrInvalidParams
	}

	previewAccessURL, cleanupRuntime, err := s.prepareCheckerPreviewAccessURL(
		ctx,
		previewService,
		previewChallengeID,
		req.AccessURL,
		req.PreviewFlag,
	)
	if err != nil {
		return nil, err
	}

	if s.roundManager == nil {
		return nil, errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}

	preview, err := s.roundManager.PreviewServiceCheck(ctx, contestports.AWDServicePreviewRequest{
		ServiceID:     previewServiceID,
		ChallengeID:   previewChallengeID,
		CheckerType:   checkerType,
		CheckerConfig: checkerConfig,
		AccessURL:     previewAccessURL,
		PreviewFlag:   req.PreviewFlag,
	})
	if cleanupErr := s.cleanupCheckerPreviewRuntime(ctx, cleanupRuntime, err); cleanupErr != nil {
		return nil, cleanupErr
	}
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.AWDCheckerPreviewResp{
		CheckerType:   preview.CheckerType,
		ServiceStatus: preview.ServiceStatus,
		CheckResult:   contestdomain.ParseAWDCheckResult(preview.CheckResult),
		PreviewContext: dto.AWDCheckerPreviewContextResp{
			ServiceID:   preview.PreviewContext.ServiceID,
			AccessURL:   preview.PreviewContext.AccessURL,
			PreviewFlag: preview.PreviewContext.PreviewFlag,
			RoundNumber: preview.PreviewContext.RoundNumber,
			TeamID:      preview.PreviewContext.TeamID,
			ChallengeID: preview.PreviewContext.ChallengeID,
		},
	}
	previewToken, err := storeAWDCheckerPreviewToken(ctx, s.redis, contestID, previewServiceID, previewChallengeID, checkerType, checkerConfig, resp)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	resp.PreviewToken = previewToken
	return resp, nil
}
