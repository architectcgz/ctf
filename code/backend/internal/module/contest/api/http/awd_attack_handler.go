package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) CreateAttackLog(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.CreateAWDAttackLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateAttackLog(c.Request.Context(), contestID, roundID, contestRequestMapper.ToCreateAttackLogInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func awdAttackLogResultsToDTO(results []contestqry.AWDAttackLogResult) []*dto.AWDAttackLogResp {
	return contestRequestMapper.ToAWDAttackLogResps(results)
}

func (h *AWDHandler) SubmitAttack(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	contestID := c.GetInt64("id")
	serviceID := c.GetInt64("sid")

	var req dto.SubmitAWDAttackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.SubmitAttack(c.Request.Context(), userID, contestID, serviceID, contestRequestMapper.ToSubmitAttackInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListAttackLogs(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.ListAttackLogs(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdAttackLogResultsToDTO(resp))
}
