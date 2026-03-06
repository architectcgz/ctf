package practice

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SubmitFlag 提交 Flag
// @Summary 提交 Flag
// @Tags Practice
// @Accept json
// @Produce json
// @Param id path int true "靶场 ID"
// @Param req body dto.SubmitFlagReq true "Flag 提交请求"
// @Success 200 {object} response.Response{data=dto.SubmissionResp}
// @Router /api/v1/challenges/{id}/submit [post]
func (h *Handler) SubmitFlag(c *gin.Context) {
	var req dto.SubmitFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FromError(c, err)
		return
	}

	challengeID := c.GetInt64("challenge_id")
	userID := c.GetInt64("user_id")

	resp, err := h.service.SubmitFlag(userID, challengeID, req.Flag)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetProgress 获取个人解题进度
// @Summary 获取个人解题进度
// @Tags Practice
// @Produce json
// @Success 200 {object} response.Response{data=dto.ProgressResp}
// @Router /api/v1/users/me/progress [get]
func (h *Handler) GetProgress(c *gin.Context) {
	userID := c.GetInt64("user_id")

	resp, err := h.service.GetProgress(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetTimeline 获取解题时间线
// @Summary 获取解题时间线
// @Tags Practice
// @Produce json
// @Param limit query int false "返回记录数" default(100)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response{data=dto.TimelineResp}
// @Router /api/v1/users/me/timeline [get]
func (h *Handler) GetTimeline(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
		Offset int `form:"offset" binding:"omitempty,min=0"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FromError(c, err)
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
