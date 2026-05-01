package http

import (
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
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

	var readinessSnapshot *contestqry.AWDReadinessResult
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

	requestCtx, gateTrace := prepareAWDReadinessGateTrace(c.Request.Context(), readinessSnapshot)
	resp, err := h.commands.UpdateContest(requestCtx, id, &req)
	writeAWDReadinessAuditPayload(c, contestdomain.AWDReadinessActionStartContest, req.OverrideReason, readinessSnapshot, gateTrace, err)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
