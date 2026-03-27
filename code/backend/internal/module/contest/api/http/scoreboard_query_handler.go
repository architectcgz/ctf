package http

import (
	"strconv"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetScoreboard(c *gin.Context) {
	h.getScoreboard(c, false)
}

func (h *Handler) GetLiveScoreboard(c *gin.Context) {
	h.getScoreboard(c, true)
}

func (h *Handler) getScoreboard(c *gin.Context, live bool) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	page := 1
	if value := c.Query("page"); value != "" {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 20
	if value := c.Query("page_size"); value != "" {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	var scoreboard *dto.ScoreboardResp
	if live {
		scoreboard, err = h.scoreboardQueries.GetLiveScoreboard(c.Request.Context(), contestID, page, pageSize)
	} else {
		scoreboard, err = h.scoreboardQueries.GetScoreboard(c.Request.Context(), contestID, page, pageSize)
	}
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, scoreboard)
}
