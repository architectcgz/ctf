package http

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type awdReadinessQueryService interface {
	GetReadiness(ctx context.Context, contestID int64) (*dto.AWDReadinessResp, error)
}

func loadAWDReadinessAuditSnapshot(ctx context.Context, queries awdReadinessQueryService, contestID int64, forceOverride *bool) (*dto.AWDReadinessResp, error) {
	if queries == nil || forceOverride == nil || !*forceOverride {
		return nil, nil
	}
	return queries.GetReadiness(ctx, contestID)
}

func writeAWDReadinessAuditPayload(c *gin.Context, gateAction string, overrideReason *string, snapshot *dto.AWDReadinessResp, err error) {
	if c == nil || snapshot == nil || snapshot.BlockingCount <= 0 || !hasNonBlankOverrideReason(overrideReason) || isAWDReadinessBlocked(err) {
		return
	}
	middleware.SetAWDReadinessAuditPayload(c, middleware.BuildAWDReadinessAuditPayload(gateAction, overrideReason, snapshot, err))
}

func shouldPrepareUpdateContestReadinessAudit(contest *dto.ContestResp, req *dto.UpdateContestReq) bool {
	if contest == nil || req == nil || req.ForceOverride == nil || !*req.ForceOverride || req.Status == nil {
		return false
	}
	return contest.Mode == model.ContestModeAWD &&
		contest.Status != model.ContestStatusRunning &&
		*req.Status == model.ContestStatusRunning
}

func isAWDReadinessBlocked(err error) bool {
	if err == nil {
		return false
	}
	var appErr *errcode.AppError
	return errors.As(err, &appErr) && appErr.Code == errcode.ErrAWDReadinessBlocked.Code
}

func hasNonBlankOverrideReason(reason *string) bool {
	return reason != nil && strings.TrimSpace(*reason) != ""
}
