package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) GetTrafficSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.GetTrafficSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListTrafficEvents(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.ListAWDTrafficEventsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.queries.ListTrafficEvents(c.Request.Context(), contestID, roundID, listAWDTrafficEventsInputFromDTO(&req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdTrafficEventPageResultToDTO(resp))
}

func listAWDTrafficEventsInputFromDTO(req *dto.ListAWDTrafficEventsReq) *contestqry.ListAWDTrafficEventsInput {
	if req == nil {
		return nil
	}
	return &contestqry.ListAWDTrafficEventsInput{
		AttackerTeamID: req.AttackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ServiceID:      req.ServiceID,
		AWDChallengeID: req.AWDChallengeID,
		StatusGroup:    req.StatusGroup,
		PathKeyword:    req.PathKeyword,
		Page:           req.Page,
		Size:           req.Size,
	}
}

func awdTrafficEventPageResultToDTO(result *contestqry.AWDTrafficEventPageResult) *dto.AWDTrafficEventPageResp {
	if result == nil {
		return nil
	}
	items := make([]*dto.AWDTrafficEventResp, 0, len(result.List))
	for i := range result.List {
		item := result.List[i]
		items = append(items, &dto.AWDTrafficEventResp{
			ID:                item.ID,
			ContestID:         item.ContestID,
			RoundID:           item.RoundID,
			AttackerTeamID:    item.AttackerTeamID,
			AttackerTeam:      item.AttackerTeam,
			AttackerTeamName:  item.AttackerTeamName,
			VictimTeamID:      item.VictimTeamID,
			VictimTeam:        item.VictimTeam,
			VictimTeamName:    item.VictimTeamName,
			ServiceID:         item.ServiceID,
			AWDChallengeID:    item.AWDChallengeID,
			AWDChallengeTitle: item.AWDChallengeTitle,
			Method:            item.Method,
			Path:              item.Path,
			StatusCode:        item.StatusCode,
			StatusGroup:       item.StatusGroup,
			IsError:           item.IsError,
			Source:            item.Source,
			OccurredAt:        item.OccurredAt,
		})
	}
	return &dto.AWDTrafficEventPageResp{
		List:     items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}
}
