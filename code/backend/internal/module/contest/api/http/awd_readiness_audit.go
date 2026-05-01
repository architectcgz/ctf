package http

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

type awdReadinessQueryService interface {
	GetReadiness(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error)
}

func loadAWDReadinessAuditSnapshot(ctx context.Context, queries awdReadinessQueryService, contestID int64, forceOverride *bool) (*contestqry.AWDReadinessResult, error) {
	if queries == nil || forceOverride == nil || !*forceOverride {
		return nil, nil
	}
	return queries.GetReadiness(ctx, contestID)
}

func prepareAWDReadinessGateTrace(ctx context.Context, snapshot *contestqry.AWDReadinessResult) (context.Context, *contestcmd.AWDReadinessGateTrace) {
	if snapshot == nil {
		return ctx, nil
	}
	return contestcmd.WithAWDReadinessGateTrace(ctx)
}

func writeAWDReadinessAuditPayload(c *gin.Context, gateAction string, overrideReason *string, snapshot *contestqry.AWDReadinessResult, trace *contestcmd.AWDReadinessGateTrace, err error) {
	if c == nil || snapshot == nil || trace == nil || !trace.Allowed() || snapshot.BlockingCount <= 0 || !hasNonBlankOverrideReason(overrideReason) || isAWDReadinessBlocked(err) {
		return
	}
	middleware.SetAWDReadinessAuditPayload(c, middleware.BuildAWDReadinessAuditPayload(gateAction, overrideReason, awdReadinessResultToDTO(snapshot), err))
}

func shouldPrepareUpdateContestReadinessAudit(contest *contestqry.ContestResult, req *dto.UpdateContestReq) bool {
	if contest == nil || req == nil || req.ForceOverride == nil || !*req.ForceOverride || req.Status == nil {
		return false
	}
	return contestdomain.ShouldGateAWDContestStart(contest.Mode, contest.Status, req.Status)
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
