package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type TagHandler struct {
	commands tagCommandService
	queries  tagQueryService
}

type tagCommandService interface {
	CreateTag(ctx context.Context, req *dto.CreateTagReq) (*dto.TagResp, error)
	AttachTags(ctx context.Context, challengeID int64, tagIDs []int64) error
	DetachTags(ctx context.Context, challengeID int64, tagIDs []int64) error
}

type tagQueryService interface {
	ListTags(ctx context.Context, tagType string) ([]*dto.TagResp, error)
}

func NewTagHandler(commands tagCommandService, queries tagQueryService) *TagHandler {
	return &TagHandler{commands: commands, queries: queries}
}

func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateTag(c.Request.Context(), &req)
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

	result, err := h.queries.ListTags(c.Request.Context(), query.Type)
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

	if err := h.commands.AttachTags(c.Request.Context(), id, req.TagIDs); err != nil {
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

	if err := h.commands.DetachTags(c.Request.Context(), id, req.TagIDs); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}
