package practice

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// StartChallenge 启动靶机实例
// POST /api/v1/challenges/:id/instances
func (h *Handler) StartChallenge(c *gin.Context) {
	userID := c.GetInt64("user_id")
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.StartChallenge(userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// GetInstance 获取实例详情
// GET /api/v1/instances/:id
func (h *Handler) GetInstance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.GetInstance(instanceID, userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// ListUserInstances 获取我的实例列表
// GET /api/v1/instances
func (h *Handler) ListUserInstances(c *gin.Context) {
	userID := c.GetInt64("user_id")

	instances, err := h.service.ListUserInstances(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instances)
}
