package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	"ctf-platform/pkg/response"
)

type skillProfileService interface {
	GetSkillProfile(ctx context.Context, userID int64) (*dto.SkillProfileResp, error)
	GetStudentSkillProfile(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.SkillProfileResp, error)
}

type recommendationProvider interface {
	Recommend(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error)
}

type Handler struct {
	service               skillProfileService
	recommendationService recommendationProvider
}

func NewHandler(service skillProfileService, recommendationService recommendationProvider) *Handler {
	return &Handler{
		service:               service,
		recommendationService: recommendationService,
	}
}

// GetMySkillProfile 获取我的能力画像
func (h *Handler) GetMySkillProfile(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	profile, err := h.service.GetSkillProfile(c.Request.Context(), userID)
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

	var req assessmentqry.RecommendationQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.recommendationService.Recommend(c.Request.Context(), userID, req.Limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}
