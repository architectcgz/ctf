package teacher

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListClasses(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)

	items, err := h.service.ListClasses(c.Request.Context(), currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, items)
}

func (h *Handler) ListClassStudents(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherStudentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, err := h.service.ListClassStudents(c.Request.Context(), currentUser.UserID, currentUser.Role, c.Param("name"), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, items)
}

func (h *Handler) GetStudentProgress(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学员ID")
		return
	}

	progress, err := h.service.GetStudentProgress(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, progress)
}

func (h *Handler) GetStudentRecommendations(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学员ID")
		return
	}

	var req struct {
		Limit int `form:"limit" binding:"omitempty,min=1,max=50"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, err := h.service.GetStudentRecommendations(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID, req.Limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, items)
}
