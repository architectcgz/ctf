package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type TagHandler struct {
	service tagService
}

type tagService interface {
	CreateTag(req *dto.CreateTagReq) (*dto.TagResp, error)
	ListTags(tagType string) ([]*dto.TagResp, error)
	AttachTags(challengeID int64, tagIDs []int64) error
	DetachTags(challengeID int64, tagIDs []int64) error
}

func NewTagHandler(service tagService) *TagHandler {
	return &TagHandler{service: service}
}

func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateTag(&req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *TagHandler) ListTags(c *gin.Context) {
	var query dto.TagQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.service.ListTags(query.Type)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *TagHandler) AttachTags(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的靶场 ID")
		return
	}

	var req dto.AttachTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.AttachTags(id, req.TagIDs); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *TagHandler) DetachTags(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的靶场 ID")
		return
	}

	var req dto.AttachTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.DetachTags(id, req.TagIDs); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}
