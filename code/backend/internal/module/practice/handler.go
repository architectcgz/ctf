package practice

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
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
	userID := authctx.MustCurrentUser(c).UserID
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
	userID := authctx.MustCurrentUser(c).UserID
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
	userID := authctx.MustCurrentUser(c).UserID

	instances, err := h.service.ListUserInstances(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instances)
}

// SubmitFlag 提交 Flag
// POST /api/v1/challenges/:id/submit
func (h *Handler) SubmitFlag(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.SubmitFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.SubmitFlag(userID, challengeID, req.Flag)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetProgress 获取个人解题进度
// GET /api/v1/users/me/progress
func (h *Handler) GetProgress(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	resp, err := h.service.GetProgress(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetTimeline 获取解题时间线
// GET /api/v1/users/me/timeline
func (h *Handler) GetTimeline(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	var req struct {
		Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
		Offset int `form:"offset" binding:"omitempty,min=0"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	resp, err := h.service.GetTimeline(userID, req.Limit, req.Offset)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
