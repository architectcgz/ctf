package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) GetTrafficSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.GetTrafficSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestRequestMapper.ToAWDTrafficSummaryRespPtr(resp))
}

func (h *AWDHandler) ListTrafficEvents(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.ListAWDTrafficEventsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.queries.ListTrafficEvents(c.Request.Context(), contestID, roundID, contestRequestMapper.ToListAWDTrafficEventsInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestRequestMapper.ToAWDTrafficEventPageRespPtr(resp))
}
