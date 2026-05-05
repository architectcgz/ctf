package http

import (
	"strconv"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ChallengeHandler) RemoveChallenge(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的题目 ID")
		return
	}

	if err := h.commands.RemoveChallengeFromContest(c.Request.Context(), contestID, challengeID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ChallengeHandler) UpdatePoints(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的题目 ID")
		return
	}

	var req dto.UpdateContestChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.commands.UpdateChallenge(c.Request.Context(), contestID, challengeID, contestRequestMapper.ToUpdateContestChallengeInput(req)); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}
