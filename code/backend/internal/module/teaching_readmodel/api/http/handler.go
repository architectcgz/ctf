package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	teachingreadmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service teachingreadmodelqueries.Service
}

func NewHandler(service teachingreadmodelqueries.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListClasses(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherClassQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, total, page, pageSize, err := h.service.ListClasses(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Page(c, items, total, page, pageSize)
}

func (h *Handler) ListStudents(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherStudentDirectoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, total, page, pageSize, err := h.service.ListStudents(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Page(c, items, total, page, pageSize)
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

func (h *Handler) GetClassSummary(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)

	summary, err := h.service.GetClassSummary(c.Request.Context(), currentUser.UserID, currentUser.Role, c.Param("name"))
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, summary)
}

func (h *Handler) GetClassTrend(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)

	trend, err := h.service.GetClassTrend(c.Request.Context(), currentUser.UserID, currentUser.Role, c.Param("name"))
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, trend)
}

func (h *Handler) GetClassReview(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)

	review, err := h.service.GetClassReview(c.Request.Context(), currentUser.UserID, currentUser.Role, c.Param("name"))
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, review)
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

	recommendations, err := h.service.GetStudentRecommendations(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID, req.Limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, recommendations)
}

func (h *Handler) GetStudentTimeline(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学员ID")
		return
	}

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

	timeline, err := h.service.GetStudentTimeline(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID, req.Limit, req.Offset)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, timeline)
}

func (h *Handler) GetStudentEvidence(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学员ID")
		return
	}

	var req dto.TeacherEvidenceQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	evidence, err := h.service.GetStudentEvidence(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, evidence)
}

func (h *Handler) GetStudentAttackSessions(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || studentID <= 0 {
		response.InvalidParams(c, "无效的学员ID")
		return
	}

	var req dto.TeacherAttackSessionQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	sessions, err := h.service.GetStudentAttackSessions(c.Request.Context(), currentUser.UserID, currentUser.Role, studentID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, sessions)
}
