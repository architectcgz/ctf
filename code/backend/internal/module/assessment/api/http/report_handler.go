package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	assessmentcommands "ctf-platform/internal/module/assessment/application/commands"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

type reportService interface {
	CreatePersonalReport(ctx context.Context, userID int64, req assessmentcommands.CreatePersonalReportInput) (*assessmentcommands.ReportExportData, error)
	CreateClassReport(ctx context.Context, requesterID int64, req assessmentcommands.CreateClassReportInput) (*assessmentcommands.ReportExportData, error)
	CreateContestExport(ctx context.Context, requesterID, contestID int64, req assessmentcommands.CreateContestExportInput) (*assessmentcommands.ReportExportData, error)
	CreateStudentReviewArchive(ctx context.Context, requesterID, studentID int64, req assessmentcommands.CreateStudentReviewArchiveInput) (*assessmentcommands.ReportExportData, error)
	GetStudentReviewArchive(ctx context.Context, requesterID, studentID int64) (*assessmentcommands.ReviewArchiveData, error)
	GetDownload(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error)
	GetStatus(ctx context.Context, reportID, requesterID int64, role string) (*assessmentcommands.ReportExportData, error)
}

type ReportHandler struct {
	service reportService
}

func NewReportHandler(service reportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) CreatePersonalReport(c *gin.Context) {
	var req CreatePersonalReportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.CreatePersonalReport(c.Request.Context(), currentUser.UserID, assessmentRequestMapper.ToCreatePersonalReportInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *ReportHandler) CreateClassReport(c *gin.Context) {
	var req CreateClassReportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.CreateClassReport(c.Request.Context(), currentUser.UserID, assessmentRequestMapper.ToCreateClassReportInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *ReportHandler) CreateContestExport(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的赛事ID")
		return
	}

	var req CreateContestExportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.CreateContestExport(c.Request.Context(), currentUser.UserID, contestID, assessmentRequestMapper.ToCreateContestExportInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *ReportHandler) CreateStudentReviewArchive(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学生ID")
		return
	}

	var req CreateStudentReviewArchiveReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.CreateStudentReviewArchive(c.Request.Context(), currentUser.UserID, studentID, assessmentRequestMapper.ToCreateStudentReviewArchiveInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *ReportHandler) GetStudentReviewArchive(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学生ID")
		return
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.GetStudentReviewArchive(c.Request.Context(), currentUser.UserID, studentID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *ReportHandler) DownloadReport(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || reportID <= 0 {
		response.InvalidParams(c, "无效的报告ID")
		return
	}

	currentUser := authctx.MustCurrentUser(c)
	download, err := h.service.GetDownload(c.Request.Context(), reportID, currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}
	defer download.Reader.Close()

	c.Header("Content-Type", download.ContentType)
	c.Header("Content-Disposition", `attachment; filename="`+download.FileName+`"`)
	c.DataFromReader(http.StatusOK, download.Size, download.ContentType, download.Reader, nil)
}

func (h *ReportHandler) GetReportStatus(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || reportID <= 0 {
		response.InvalidParams(c, "无效的报告ID")
		return
	}

	currentUser := authctx.MustCurrentUser(c)
	report, err := h.service.GetStatus(c.Request.Context(), reportID, currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, report)
}
