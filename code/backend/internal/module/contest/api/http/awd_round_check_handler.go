package http

import (
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) RunCurrentRoundChecks(c *gin.Context) {
	contestID := c.GetInt64("id")
	req := dto.RunCurrentAWDCheckerReq{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}
	input := contestRequestMapper.ToRunCurrentRoundChecksInput(req)

	readinessSnapshot, err := loadAWDReadinessAuditSnapshot(c.Request.Context(), h.queries, contestID, input.ForceOverride)
	if err != nil {
		response.FromError(c, err)
		return
	}

	requestCtx, gateTrace := prepareAWDReadinessGateTrace(c.Request.Context(), readinessSnapshot)
	resp, err := h.commands.RunCurrentRoundChecks(requestCtx, contestID, input)
	writeAWDReadinessAuditPayload(c, contestdomain.AWDReadinessActionRunCurrentRoundCheck, input.OverrideReason, readinessSnapshot, gateTrace, err)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) RunRoundChecks(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.commands.RunRoundChecks(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) PreviewChecker(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.PreviewAWDCheckerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	input := contestRequestMapper.ToPreviewCheckerInput(req)

	requestCtx := contestcmd.WithAWDPreviewRequester(c.Request.Context(), authctx.MustCurrentUser(c).UserID)
	resp, err := h.commands.PreviewChecker(requestCtx, contestID, input)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
