package http

import (
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ParticipationHandler) ListRegistrations(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	var query dto.ContestRegistrationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}
	items, err := h.queries.ListRegistrations(c.Request.Context(), contestID, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *ParticipationHandler) ListAnnouncements(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	items, err := h.queries.ListAnnouncements(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestAnnouncementResultsToDTO(items))
}

func (h *ParticipationHandler) GetMyProgress(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	item, err := h.queries.GetMyProgress(c.Request.Context(), contestID, authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, participationProgressResultToDTO(item))
}

func contestAnnouncementResultsToDTO(items []*contestqry.ContestAnnouncementResult) []*dto.ContestAnnouncementResp {
	result := make([]*dto.ContestAnnouncementResp, 0, len(items))
	for _, item := range items {
		if item == nil {
			result = append(result, nil)
			continue
		}
		result = append(result, &dto.ContestAnnouncementResp{
			ID:        item.ID,
			Title:     item.Title,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		})
	}
	return result
}

func participationProgressResultToDTO(item *contestqry.ParticipationProgressResult) *dto.ContestMyProgressResp {
	if item == nil {
		return nil
	}
	result := &dto.ContestMyProgressResp{
		ContestID: item.ContestID,
		TeamID:    item.TeamID,
		Solved:    make([]*dto.ContestSolvedProgressItem, 0, len(item.Solved)),
	}
	for _, solved := range item.Solved {
		if solved == nil {
			result.Solved = append(result.Solved, nil)
			continue
		}
		result.Solved = append(result.Solved, &dto.ContestSolvedProgressItem{
			ContestChallengeID: solved.ContestChallengeID,
			SolvedAt:           solved.SolvedAt,
			PointsEarned:       solved.PointsEarned,
		})
	}
	return result
}
