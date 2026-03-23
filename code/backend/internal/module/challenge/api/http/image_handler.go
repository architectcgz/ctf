package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type ImageHandler struct {
	service imageService
}

type imageService interface {
	CreateImageWithContext(ctx context.Context, req *dto.CreateImageReq) (*dto.ImageResp, error)
	GetImage(id int64) (*dto.ImageResp, error)
	ListImages(query *dto.ImageQuery) (*dto.PageResult, error)
	UpdateImage(id int64, req *dto.UpdateImageReq) error
	DeleteImage(id int64) error
}

func NewImageHandler(service imageService) *ImageHandler {
	return &ImageHandler{service: service}
}

func (h *ImageHandler) CreateImage(c *gin.Context) {
	var req dto.CreateImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateImageWithContext(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的镜像 ID")
		return
	}

	resp, err := h.service.GetImage(id)
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

	result, err := h.service.ListImages(&query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *ImageHandler) UpdateImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的镜像 ID")
		return
	}

	var req dto.UpdateImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.UpdateImage(id, &req); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的镜像 ID")
		return
	}

	if err := h.service.DeleteImage(id); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}
