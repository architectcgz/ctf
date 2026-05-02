package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) UpsertServiceCheck(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.UpsertAWDServiceCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.UpsertServiceCheck(c.Request.Context(), contestID, roundID, contestRequestMapper.ToUpsertServiceCheckInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListServices(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.ListServices(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestRequestMapper.ToAWDTeamServiceResps(resp))
}
