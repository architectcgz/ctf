package http

import (
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	var req dto.CreateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := authctx.MustCurrentUser(c).UserID
	teamResp, err := h.commands.CreateTeam(c.Request.Context(), contestID, userID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, teamResp)
}

func (h *TeamHandler) JoinTeam(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	teamID, err := strconv.ParseInt(c.Param("tid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的队伍ID")
		return
	}

	userID := authctx.MustCurrentUser(c).UserID
	teamResp, err := h.commands.JoinTeam(c.Request.Context(), contestID, userID, teamID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, teamResp)
}
