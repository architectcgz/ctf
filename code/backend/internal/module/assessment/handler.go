package assessment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service               *Service
	recommendationService *RecommendationService
}

func NewHandler(service *Service, recommendationService *RecommendationService) *Handler {
	return &Handler{
		service:               service,
		recommendationService: recommendationService,
	}
}

// GetMySkillProfile 获取我的能力画像
func (h *Handler) GetMySkillProfile(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

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

	currentUser := authctx.MustCurrentUser(c)
	profile, err := h.service.GetStudentSkillProfile(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, profile)
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	var req struct {
		Limit int `form:"limit"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.recommendationService.Recommend(userID, req.Limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}
