package http

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateContest(c *gin.Context) {
	var req dto.CreateContestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateContest(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) UpdateContest(c *gin.Context) {
	id := c.GetInt64("id")
	var req dto.UpdateContestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var readinessSnapshot *dto.AWDReadinessResp
	if h.readinessQueries != nil && h.queries != nil {
		contest, err := h.queries.GetContest(c.Request.Context(), id)
		if err != nil {
			response.FromError(c, err)
			return
		}
		if shouldPrepareUpdateContestReadinessAudit(contest, &req) {
			readinessSnapshot, err = loadAWDReadinessAuditSnapshot(c.Request.Context(), h.readinessQueries, id, req.ForceOverride)
			if err != nil {
				response.FromError(c, err)
				return
			}
		}
	}

	resp, err := h.commands.UpdateContest(c.Request.Context(), id, &req)
	writeAWDReadinessAuditPayload(c, "start_contest", req.OverrideReason, readinessSnapshot, err)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
