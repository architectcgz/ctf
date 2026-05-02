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
	items, err := h.queries.ListRegistrations(c.Request.Context(), contestID, contestqry.ContestRegistrationQueryInput{
		Status: query.Status,
		Page:   query.Page,
		Size:   query.Size,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestRegistrationPageResultToDTO(items))
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
	return contestRequestMapper.ToContestAnnouncementResps(items)
}

func participationProgressResultToDTO(item *contestqry.ParticipationProgressResult) *dto.ContestMyProgressResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToContestMyProgressResp(*item)
	return &mapped
}

func contestRegistrationPageResultToDTO(item *contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult]) *dto.PageResult[*dto.ContestRegistrationResp] {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToRegistrationPageResp(*item)
	return &mapped
}
