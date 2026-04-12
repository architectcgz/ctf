package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type teacherAWDReviewService interface {
	ListContests(ctx context.Context, requesterID int64) (*dto.TeacherAWDReviewContestListResp, error)
	GetContestArchive(ctx context.Context, requesterID, contestID int64, req *dto.GetTeacherAWDReviewArchiveReq) (*dto.TeacherAWDReviewArchiveResp, error)
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
	response.Error(c, errcode.New(errcode.ErrServiceUnavailable.Code, "教师 AWD 复盘归档导出暂未实现", http.StatusNotImplemented))
}

func (h *TeacherAWDReviewHandler) ExportReport(c *gin.Context) {
	response.Error(c, errcode.New(errcode.ErrServiceUnavailable.Code, "教师 AWD 复盘报告导出暂未实现", http.StatusNotImplemented))
}
