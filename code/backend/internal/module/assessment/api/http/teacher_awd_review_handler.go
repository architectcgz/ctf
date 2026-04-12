package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type teacherAWDReviewService interface {
	ListContests(ctx context.Context, requesterID int64) (*dto.TeacherAWDReviewContestListResp, error)
	GetContestArchive(ctx context.Context, requesterID, contestID int64, req *dto.GetTeacherAWDReviewArchiveReq) (*dto.TeacherAWDReviewArchiveResp, error)
	CreateTeacherAWDReviewArchive(ctx context.Context, requesterID, contestID int64, req *dto.CreateTeacherAWDReviewExportReq) (*dto.ReportExportData, error)
	CreateTeacherAWDReviewReport(ctx context.Context, requesterID, contestID int64, req *dto.CreateTeacherAWDReviewExportReq) (*dto.ReportExportData, error)
}

type TeacherAWDReviewHandler struct {
	service teacherAWDReviewService
}

func NewTeacherAWDReviewHandler(service teacherAWDReviewService) *TeacherAWDReviewHandler {
	return &TeacherAWDReviewHandler{service: service}
}

func (h *TeacherAWDReviewHandler) ListReviews(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.ListContests(c.Request.Context(), currentUser.UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TeacherAWDReviewHandler) GetReview(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	contestID := c.GetInt64("id")

	var req dto.GetTeacherAWDReviewArchiveReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.GetContestArchive(c.Request.Context(), currentUser.UserID, contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TeacherAWDReviewHandler) ExportArchive(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	contestID := c.GetInt64("id")

	var req dto.CreateTeacherAWDReviewExportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	resp, err := h.service.CreateTeacherAWDReviewArchive(c.Request.Context(), currentUser.UserID, contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TeacherAWDReviewHandler) ExportReport(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	contestID := c.GetInt64("id")

	var req dto.CreateTeacherAWDReviewExportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	resp, err := h.service.CreateTeacherAWDReviewReport(c.Request.Context(), currentUser.UserID, contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
