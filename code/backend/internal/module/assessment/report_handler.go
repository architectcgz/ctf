package assessment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type ReportHandler struct {
	service *ReportService
}

func NewReportHandler(service *ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) CreatePersonalReport(c *gin.Context) {
	var req dto.CreatePersonalReportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.CreatePersonalReport(c.Request.Context(), currentUser.UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *ReportHandler) CreateClassReport(c *gin.Context) {
	var req dto.CreateClassReportReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, err)
			return
		}
	}

	currentUser := authctx.MustCurrentUser(c)
	resp, err := h.service.CreateClassReport(c.Request.Context(), currentUser.UserID, &req)
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

	c.Header("Content-Type", download.ContentType)
	c.FileAttachment(download.Path, download.FileName)
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
