package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
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
	response.Success(c, awdTrafficSummaryResultToDTO(resp))
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
	response.Success(c, awdTrafficEventPageResultToDTO(resp))
}

func awdTrafficEventPageResultToDTO(result *contestqry.AWDTrafficEventPageResult) *dto.AWDTrafficEventPageResp {
	if result == nil {
		return nil
	}
	mapped := contestRequestMapper.ToAWDTrafficEventPageResp(*result)
	return &mapped
}

func awdTrafficSummaryResultToDTO(item *contestqry.AWDTrafficSummaryResult) *dto.AWDTrafficSummaryResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToAWDTrafficSummaryResp(*item)
	return &mapped
}
