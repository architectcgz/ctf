package http

import (
	"context"
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type teamCommandService interface {
	CreateTeam(ctx context.Context, contestID, captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error)
	JoinTeam(ctx context.Context, contestID, userID, teamID int64) (*dto.TeamResp, error)
	LeaveTeam(ctx context.Context, contestID, userID, teamID int64) error
	DismissTeam(ctx context.Context, contestID, captainID, teamID int64) error
	KickMember(ctx context.Context, contestID, captainID, teamID, memberUserID int64) error
}

type teamQueryService interface {
	GetTeamInfo(teamID int64) (*dto.TeamResp, []*dto.TeamMemberResp, error)
	GetMyTeam(ctx context.Context, contestID, userID int64) (map[string]any, error)
	ListTeams(ctx context.Context, contestID int64) ([]*dto.TeamResp, error)
}

type TeamHandler struct {
	commands teamCommandService
	queries  teamQueryService
}

func NewTeamHandler(commands teamCommandService, queries teamQueryService) *TeamHandler {
	return &TeamHandler{commands: commands, queries: queries}
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

func (h *TeamHandler) GetTeamInfo(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("tid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的队伍ID")
		return
	}

	teamResp, members, err := h.queries.GetTeamInfo(teamID)
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

	teams, err := h.queries.ListTeams(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, teams)
}

func (h *TeamHandler) GetMyTeam(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	userID := authctx.MustCurrentUser(c).UserID
	team, err := h.queries.GetMyTeam(c.Request.Context(), contestID, userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, team)
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
