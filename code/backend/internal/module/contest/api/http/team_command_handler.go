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

func (h *TeamHandler) LeaveTeam(c *gin.Context) {
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
	if err := h.commands.LeaveTeam(c.Request.Context(), contestID, userID, teamID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "退出成功"})
}

func (h *TeamHandler) DismissTeam(c *gin.Context) {
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
	if err := h.commands.DismissTeam(c.Request.Context(), contestID, userID, teamID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "解散成功"})
}

func (h *TeamHandler) KickMember(c *gin.Context) {
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
	memberUserID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	userID := authctx.MustCurrentUser(c).UserID
	if err := h.commands.KickMember(c.Request.Context(), contestID, userID, teamID, memberUserID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "移除成功"})
}
