package http

import (
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

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
