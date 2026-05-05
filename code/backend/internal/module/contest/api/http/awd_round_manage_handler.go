package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) CreateRound(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateAWDRoundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	input := contestRequestMapper.ToCreateAWDRoundInput(req)

	readinessSnapshot, err := loadAWDReadinessAuditSnapshot(c.Request.Context(), h.queries, contestID, input.ForceOverride)
	if err != nil {
		response.FromError(c, err)
		return
	}

	requestCtx, gateTrace := prepareAWDReadinessGateTrace(c.Request.Context(), readinessSnapshot)
	resp, err := h.commands.CreateRound(requestCtx, contestID, input)
	writeAWDReadinessAuditPayload(c, contestdomain.AWDReadinessActionCreateRound, input.OverrideReason, readinessSnapshot, gateTrace, err)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListRounds(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.queries.ListRounds(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestRequestMapper.ToAWDRoundResps(resp))
}
