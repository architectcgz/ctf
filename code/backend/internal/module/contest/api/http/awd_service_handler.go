package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) UpsertServiceCheck(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.UpsertAWDServiceCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.UpsertServiceCheck(c.Request.Context(), contestID, roundID, contestRequestMapper.ToUpsertServiceCheckInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func awdTeamServiceResultsToDTO(results []contestqry.AWDTeamServiceResult) []*dto.AWDTeamServiceResp {
	resp := make([]*dto.AWDTeamServiceResp, 0, len(results))
	for i := range results {
		item := results[i]
		resp = append(resp, &dto.AWDTeamServiceResp{
			ID:                item.ID,
			RoundID:           item.RoundID,
			TeamID:            item.TeamID,
			TeamName:          item.TeamName,
			ServiceID:         item.ServiceID,
			ServiceName:       item.ServiceName,
			AWDChallengeID:    item.AWDChallengeID,
			AWDChallengeTitle: item.AWDChallengeTitle,
			ServiceStatus:     item.ServiceStatus,
			CheckResult:       item.CheckResult,
			CheckerType:       model.AWDCheckerType(item.CheckerType),
			AttackReceived:    item.AttackReceived,
			SLAScore:          item.SLAScore,
			DefenseScore:      item.DefenseScore,
			AttackScore:       item.AttackScore,
			UpdatedAt:         item.UpdatedAt,
		})
	}
	return resp
}

func (h *AWDHandler) ListServices(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.ListServices(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdTeamServiceResultsToDTO(resp))
}
