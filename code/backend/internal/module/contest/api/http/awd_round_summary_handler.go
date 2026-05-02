package http

import (
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) GetRoundSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.GetRoundSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdRoundSummaryResultToDTO(resp))
}

func awdRoundSummaryResultToDTO(item *contestqry.AWDRoundSummaryResult) *dto.AWDRoundSummaryResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToAWDRoundSummaryResp(*item)
	return &mapped
}
