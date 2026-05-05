package http

import (
	"strconv"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ChallengeHandler) AddChallenge(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	var req dto.AddContestChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.AddChallengeToContest(c.Request.Context(), contestID, contestRequestMapper.ToAddContestChallengeInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
