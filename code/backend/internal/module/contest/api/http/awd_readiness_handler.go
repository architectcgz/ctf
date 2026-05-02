package http

import (
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) GetReadiness(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.queries.GetReadiness(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdReadinessResultToDTO(resp))
}

func awdReadinessResultToDTO(result *contestqry.AWDReadinessResult) *dto.AWDReadinessResp {
	if result == nil {
		return nil
	}
	mapped := contestRequestMapper.ToAWDReadinessResp(*result)
	return &mapped
}
