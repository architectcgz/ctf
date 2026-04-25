package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

const errMsgInvalidImageID = "无效的镜像 ID"

type ImageHandler struct {
	commands imageCommandService
	queries  imageQueryService
}

type imageCommandService interface {
	CreateImage(ctx context.Context, req *dto.CreateImageReq) (*dto.ImageResp, error)
	UpdateImage(ctx context.Context, id int64, req *dto.UpdateImageReq) error
	DeleteImage(ctx context.Context, id int64) error
}

type imageQueryService interface {
	GetImage(ctx context.Context, id int64) (*dto.ImageResp, error)
	ListImages(ctx context.Context, query *dto.ImageQuery) (*dto.PageResult, error)
}

func NewImageHandler(commands imageCommandService, queries imageQueryService) *ImageHandler {
	return &ImageHandler{commands: commands, queries: queries}
}

func (h *ImageHandler) CreateImage(c *gin.Context) {
	var req dto.CreateImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateImage(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, errMsgInvalidImageID)
		return
	}

	resp, err := h.queries.GetImage(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ImageHandler) ListImages(c *gin.Context) {
	var query dto.ImageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.queries.ListImages(c.Request.Context(), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *ImageHandler) UpdateImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, errMsgInvalidImageID)
		return
	}

	var req dto.UpdateImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.commands.UpdateImage(c.Request.Context(), id, &req); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, errMsgInvalidImageID)
		return
	}

	if err := h.commands.DeleteImage(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}
