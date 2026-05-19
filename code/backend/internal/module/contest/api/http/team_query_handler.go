package http

import (
	"strconv"

	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"

	"github.com/gin-gonic/gin"
)

func (h *TeamHandler) GetTeamInfo(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("tid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的队伍ID")
		return
	}

	teamResp, members, err := h.queries.GetTeamInfo(c.Request.Context(), teamID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, TeamInfoResp{
		Team:    contestResponseMapper.ToTeamRespPtr(teamResp),
		Members: contestResponseMapper.ToTeamMemberResps(members),
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

	response.Success(c, contestResponseMapper.ToTeamResps(teams))
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
	response.Success(c, contestResponseMapper.ToMyTeamRespPtr(team))
}
