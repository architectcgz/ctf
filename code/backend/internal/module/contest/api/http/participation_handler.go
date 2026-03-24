package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type participationCommandService interface {
	RegisterContest(ctx context.Context, contestID, userID int64) error
	ReviewRegistration(ctx context.Context, contestID, registrationID, reviewerID int64, req *dto.ReviewContestRegistrationReq) (*dto.ContestRegistrationResp, error)
	CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req *dto.CreateContestAnnouncementReq) (*dto.ContestAnnouncementResp, error)
	DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error
}

type participationQueryService interface {
	ListRegistrations(ctx context.Context, contestID int64, query *dto.ContestRegistrationQuery) (*dto.PageResult, error)
	ListAnnouncements(ctx context.Context, contestID int64) ([]*dto.ContestAnnouncementResp, error)
	GetMyProgress(ctx context.Context, contestID, userID int64) (*dto.ContestMyProgressResp, error)
}

type ParticipationHandler struct {
	commands participationCommandService
	queries  participationQueryService
}

func NewParticipationHandler(commands participationCommandService, queries participationQueryService) *ParticipationHandler {
	return &ParticipationHandler{commands: commands, queries: queries}
}

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
	response.Success(c, items)
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
	response.Success(c, item)
}
