package container

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

func (h *Handler) CreateInstance(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateInstance(userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) DestroyInstance(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.DestroyInstance(instanceID, userID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ExtendInstance(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.ExtendInstance(instanceID, userID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListInstances(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	instances, err := h.service.GetUserInstances(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instances)
}

func (h *Handler) ListTeacherInstances(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherInstanceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, err := h.service.ListTeacherInstances(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, items)
}

func (h *Handler) DestroyTeacherInstance(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	instanceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.DestroyTeacherInstance(c.Request.Context(), instanceID, currentUser.UserID, currentUser.Role); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}
