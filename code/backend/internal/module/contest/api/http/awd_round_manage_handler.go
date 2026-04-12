package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) CreateRound(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateAWDRoundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	readinessSnapshot, err := loadAWDReadinessAuditSnapshot(c.Request.Context(), h.queries, contestID, req.ForceOverride)
	if err != nil {
		response.FromError(c, err)
		return
	}

	resp, err := h.commands.CreateRound(c.Request.Context(), contestID, &req)
	writeAWDReadinessAuditPayload(c, "create_round", req.OverrideReason, readinessSnapshot, err)
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
	response.Success(c, resp)
}
