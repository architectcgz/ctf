package http

import (
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
	response.Success(c, contestRequestMapper.ToAWDReadinessRespPtr(resp))
}
