package http

import (
	"strconv"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) FreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	var req dto.FreezeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.scoreboardCommand.FreezeScoreboard(c.Request.Context(), contestID, req.MinutesBeforeEnd); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "排行榜已冻结"})
}

func (h *Handler) UnfreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	if err := h.scoreboardCommand.UnfreezeScoreboard(c.Request.Context(), contestID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "排行榜已解冻"})
}
