package http

import (
	"context"
	"strconv"

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
	CreateTemplate(ctx context.Context, actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error)
	UpdateTemplate(ctx context.Context, id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error)
	DeleteTemplate(ctx context.Context, id int64) error
}

type awdServiceTemplateQueryService interface {
	GetTemplate(ctx context.Context, id int64) (*dto.AWDServiceTemplateResp, error)
	ListTemplates(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error)
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
	resp, err := h.commands.CreateTemplate(c.Request.Context(), authctx.MustCurrentUser(c).UserID, &req)
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
	resp, err := h.queries.GetTemplate(c.Request.Context(), id)
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
	resp, err := h.queries.ListTemplates(c.Request.Context(), &req)
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
	resp, err := h.commands.UpdateTemplate(c.Request.Context(), id, &req)
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
	if err := h.commands.DeleteTemplate(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
