package adminuser

import (
	"net/http"

	"github.com/gin-gonic/gin"

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

func (h *Handler) ListUsers(c *gin.Context) {
	var query dto.AdminUserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	list, total, page, size, err := h.service.ListUsers(c.Request.Context(), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, gin.H{"user": user})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID := c.GetInt64("id")
	var req dto.UpdateAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, gin.H{"user": user})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.GetInt64("id")
	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *Handler) ImportUsers(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.InvalidParams(c, "缺少导入文件")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, errcode.New(errcode.ErrInvalidParams.Code, "无法读取导入文件", errcode.ErrInvalidParams.HTTPStatus))
		return
	}
	defer file.Close()

	result, err := h.service.ImportUsers(c.Request.Context(), file)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, result)
}
