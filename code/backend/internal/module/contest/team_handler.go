package contest

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *TeamService
}

func NewTeamHandler(teamService *TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

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

	userID := c.GetInt64("user_id")
	teamResp, err := h.teamService.CreateTeam(c.Request.Context(), contestID, userID, &req)
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

	userID := c.GetInt64("user_id")
	teamResp, err := h.teamService.JoinTeam(c.Request.Context(), contestID, userID, teamID)
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

	userID := c.GetInt64("user_id")
	if err := h.teamService.LeaveTeam(c.Request.Context(), contestID, userID, teamID); err != nil {
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

	userID := c.GetInt64("user_id")
	if err := h.teamService.DismissTeam(c.Request.Context(), contestID, userID, teamID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "解散成功"})
}

func (h *TeamHandler) GetTeamInfo(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("tid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的队伍ID")
		return
	}

	teamResp, members, err := h.teamService.GetTeamInfo(teamID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{
		"team":    teamResp,
		"members": members,
	})
}

func (h *TeamHandler) ListTeams(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	teams, err := h.teamService.ListTeams(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, teams)
}
