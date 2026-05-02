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
	result := make([]*dto.TeamResp, 0, len(items))
	for _, item := range items {
		result = append(result, teamResultToDTO(item))
	}
	return result
}

func teamResultToDTO(item *contestqry.TeamResult) *dto.TeamResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToTeamResp(*item)
	return &mapped
}

func teamMemberResultsToDTO(items []*contestqry.TeamMemberResult) []*dto.TeamMemberResp {
	result := make([]*dto.TeamMemberResp, 0, len(items))
	for _, item := range items {
		if item == nil {
			result = append(result, nil)
			continue
		}
		mapped := contestRequestMapper.ToTeamMemberResp(*item)
		result = append(result, &mapped)
	}
	return result
}

func myTeamResultToDTO(item *contestqry.MyTeamResult) gin.H {
	if item == nil {
		return nil
	}
	return gin.H{
		"id":              item.ID,
		"name":            item.Name,
		"invite_code":     item.InviteCode,
		"captain_user_id": item.CaptainID,
		"members":         teamMemberResultsToDTO(item.Members),
	}
}
