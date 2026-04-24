package http

import (
	"context"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type AWDServiceTemplateHandler struct {
	commands awdServiceTemplateCommandService
	queries  awdServiceTemplateQueryService
}

type awdServiceTemplateCommandService interface {
	CreateTemplate(actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error)
	CreateTemplateWithContext(ctx context.Context, actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error)
	UpdateTemplate(id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error)
	UpdateTemplateWithContext(ctx context.Context, id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error)
	DeleteTemplate(id int64) error
	DeleteTemplateWithContext(ctx context.Context, id int64) error
	PreviewImport(actorUserID int64, fileName string, reader io.Reader) (*dto.AWDServiceTemplateImportPreviewResp, error)
	PreviewImportWithContext(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*dto.AWDServiceTemplateImportPreviewResp, error)
	ListImports(actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error)
	ListImportsWithContext(ctx context.Context, actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error)
	GetImport(actorUserID int64, id string) (*dto.AWDServiceTemplateImportPreviewResp, error)
	GetImportWithContext(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateImportPreviewResp, error)
	CommitImport(actorUserID int64, id string) (*dto.AWDServiceTemplateResp, error)
	CommitImportWithContext(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateResp, error)
}

type awdServiceTemplateQueryService interface {
	GetTemplateWithContext(ctx context.Context, id int64) (*dto.AWDServiceTemplateResp, error)
	ListTemplatesWithContext(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error)
}

func NewAWDServiceTemplateHandler(commands awdServiceTemplateCommandService, queries awdServiceTemplateQueryService) *AWDServiceTemplateHandler {
	return &AWDServiceTemplateHandler{commands: commands, queries: queries}
}

func (h *AWDServiceTemplateHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateAWDServiceTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.CreateTemplateWithContext(c.Request.Context(), authctx.MustCurrentUser(c).UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDServiceTemplateHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Service Template ID")
		return
	}
	resp, err := h.queries.GetTemplateWithContext(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDServiceTemplateHandler) ListTemplates(c *gin.Context) {
	var req dto.AWDServiceTemplateQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.queries.ListTemplatesWithContext(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDServiceTemplateHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Service Template ID")
		return
	}
	var req dto.UpdateAWDServiceTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.UpdateTemplateWithContext(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDServiceTemplateHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Service Template ID")
		return
	}
	if err := h.commands.DeleteTemplateWithContext(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *AWDServiceTemplateHandler) PreviewImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.InvalidParams(c, "缺少 AWD 题目包文件")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InvalidParams(c, "无法读取 AWD 题目包文件")
		return
	}
	defer file.Close()

	resp, err := h.commands.PreviewImportWithContext(
		c.Request.Context(),
		authctx.MustCurrentUser(c).UserID,
		fileHeader.Filename,
		file,
	)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.SuccessWithStatus(c, nethttp.StatusCreated, resp)
}

func (h *AWDServiceTemplateHandler) ListImports(c *gin.Context) {
	resp, err := h.commands.ListImportsWithContext(c.Request.Context(), authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDServiceTemplateHandler) GetImport(c *gin.Context) {
	resp, err := h.commands.GetImportWithContext(c.Request.Context(), authctx.MustCurrentUser(c).UserID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDServiceTemplateHandler) CommitImport(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.InvalidParams(c, "无效的导入 ID")
		return
	}

	resp, err := h.commands.CommitImportWithContext(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, &dto.AWDServiceTemplateImportCommitResp{Template: resp})
}
