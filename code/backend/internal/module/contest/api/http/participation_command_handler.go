package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

func (h *ParticipationHandler) RegisterContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	if err := h.commands.RegisterContest(c.Request.Context(), contestID, authctx.MustCurrentUser(c).UserID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ParticipationHandler) ReviewRegistration(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	registrationID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil || registrationID <= 0 {
		response.InvalidParams(c, "无效的报名ID")
		return
	}
	var req dto.ReviewContestRegistrationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	item, err := h.commands.ReviewRegistration(c.Request.Context(), contestID, registrationID, authctx.MustCurrentUser(c).UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *ParticipationHandler) CreateAnnouncement(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	var req dto.CreateContestAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	item, err := h.commands.CreateAnnouncement(c.Request.Context(), contestID, authctx.MustCurrentUser(c).UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *ParticipationHandler) DeleteAnnouncement(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}
	announcementID, err := strconv.ParseInt(c.Param("aid"), 10, 64)
	if err != nil || announcementID <= 0 {
		response.InvalidParams(c, "无效的公告ID")
		return
	}
	if err := h.commands.DeleteAnnouncement(c.Request.Context(), contestID, announcementID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
