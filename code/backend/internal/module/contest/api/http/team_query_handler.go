package http

import (
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

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

	response.Success(c, gin.H{
		"team":    teamResultToDTO(teamResp),
		"members": teamMemberResultsToDTO(members),
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

	response.Success(c, teamResultsToDTO(teams))
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
	response.Success(c, myTeamResultToDTO(team))
}

func teamResultsToDTO(items []*contestqry.TeamResult) []*dto.TeamResp {
	return contestRequestMapper.ToTeamResps(items)
}

func teamResultToDTO(item *contestqry.TeamResult) *dto.TeamResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToTeamResp(*item)
	return &mapped
}

func teamMemberResultsToDTO(items []*contestqry.TeamMemberResult) []*dto.TeamMemberResp {
	return contestRequestMapper.ToTeamMemberResps(items)
}

func myTeamResultToDTO(item *contestqry.MyTeamResult) *dto.MyTeamResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToMyTeamResp(*item)
	return &mapped
}
