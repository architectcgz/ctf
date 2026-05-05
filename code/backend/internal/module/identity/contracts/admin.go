package contracts

import (
	"context"
	"io"

	"ctf-platform/internal/dto"
)

type CreateUserInput struct {
	Username  string
	Password  string
	Name      string
	Email     string
	StudentNo string
	TeacherNo string
	ClassName string
	Role      string
	Status    string
}

type UpdateUserInput struct {
	Password  *string
	Name      *string
	Email     *string
	StudentNo *string
	TeacherNo *string
	ClassName *string
	Role      *string
	Status    *string
}

type AdminCommandService interface {
	CreateUser(ctx context.Context, req CreateUserInput) (*dto.AdminUserResp, error)
	UpdateUser(ctx context.Context, userID int64, req UpdateUserInput) (*dto.AdminUserResp, error)
	DeleteUser(ctx context.Context, userID int64) error
	ImportUsers(ctx context.Context, reader io.Reader) (*dto.ImportUsersResp, error)
}

type AdminQueryService interface {
	ListUsers(ctx context.Context, query *dto.AdminUserQuery) ([]dto.AdminUserResp, int64, int, int, error)
}
