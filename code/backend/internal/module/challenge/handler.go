package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateChallenge(c *gin.Context) {
	var req dto.CreateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateChallenge(&req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) UpdateChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	var req dto.UpdateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.UpdateChallenge(id, &req); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	if err := h.service.DeleteChallenge(id); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	resp, err := h.service.GetChallenge(id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) ListChallenges(c *gin.Context) {
	var query dto.ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.service.ListChallenges(&query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *Handler) PublishChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	if err := h.service.PublishChallenge(id); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListPublishedChallenges 靶场列表（学员视图）
func (h *Handler) ListPublishedChallenges(c *gin.Context) {
	var query dto.ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = int64(0)
	}

	uid, ok := userID.(int64)
	if !ok {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	result, err := h.service.ListPublishedChallenges(uid, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

// GetPublishedChallenge 靶场详情（学员视图）
func (h *Handler) GetPublishedChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = int64(0)
	}

	uid, ok := userID.(int64)
	if !ok {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	detail, err := h.service.GetPublishedChallenge(uid, id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, detail)
}
