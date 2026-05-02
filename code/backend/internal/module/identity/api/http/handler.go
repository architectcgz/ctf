package http

import (
	"context"
	"io"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type adminCommandService interface {
	CreateUser(ctx context.Context, req identitycontracts.CreateUserInput) (*dto.AdminUserResp, error)
	UpdateUser(ctx context.Context, userID int64, req identitycontracts.UpdateUserInput) (*dto.AdminUserResp, error)
	DeleteUser(ctx context.Context, userID int64) error
	ImportUsers(ctx context.Context, reader io.Reader) (*dto.ImportUsersResp, error)
}

type adminQueryService interface {
	ListUsers(ctx context.Context, query *dto.AdminUserQuery) ([]dto.AdminUserResp, int64, int, int, error)
}

type Handler struct {
	commands adminCommandService
	queries  adminQueryService
}

func NewHandler(commands adminCommandService, queries adminQueryService) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

func (h *Handler) ListUsers(c *gin.Context) {
	var query dto.AdminUserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	list, total, page, size, err := h.queries.ListUsers(c.Request.Context(), &query)
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

	user, err := h.commands.CreateUser(c.Request.Context(), createUserInputFromDTO(&req))
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

	user, err := h.commands.UpdateUser(c.Request.Context(), userID, updateUserInputFromDTO(&req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, gin.H{"user": user})
}

func createUserInputFromDTO(req *dto.CreateAdminUserReq) identitycontracts.CreateUserInput {
	if req == nil {
		return identitycontracts.CreateUserInput{}
	}
	return identitycontracts.CreateUserInput{
		Username:  req.Username,
		Password:  req.Password,
		Name:      req.Name,
		Email:     req.Email,
		StudentNo: req.StudentNo,
		TeacherNo: req.TeacherNo,
		ClassName: req.ClassName,
		Role:      req.Role,
		Status:    req.Status,
	}
}

func updateUserInputFromDTO(req *dto.UpdateAdminUserReq) identitycontracts.UpdateUserInput {
	if req == nil {
		return identitycontracts.UpdateUserInput{}
	}
	return identitycontracts.UpdateUserInput{
		Password:  req.Password,
		Name:      req.Name,
		Email:     req.Email,
		StudentNo: req.StudentNo,
		TeacherNo: req.TeacherNo,
		ClassName: req.ClassName,
		Role:      req.Role,
		Status:    req.Status,
	}
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.GetInt64("id")
	if err := h.commands.DeleteUser(c.Request.Context(), userID); err != nil {
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

	result, err := h.commands.ImportUsers(c.Request.Context(), file)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.SuccessWithStatus(c, nethttp.StatusCreated, result)
}
