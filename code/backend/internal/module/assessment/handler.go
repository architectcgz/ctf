package assessment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMySkillProfile 获取我的能力画像
func (h *Handler) GetMySkillProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	profile, err := h.service.GetSkillProfile(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, profile)
}

// GetStudentSkillProfile 教师查看学员能力画像
func (h *Handler) GetStudentSkillProfile(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的学员ID")
		return
	}

	profile, err := h.service.GetSkillProfile(studentID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, profile)
}
