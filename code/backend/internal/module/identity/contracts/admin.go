package contracts

import (
	"context"
	"io"

	"ctf-platform/internal/dto"
)

type AdminCommandService interface {
	CreateUser(ctx context.Context, req *dto.CreateAdminUserReq) (*dto.AdminUserResp, error)
	UpdateUser(ctx context.Context, userID int64, req *dto.UpdateAdminUserReq) (*dto.AdminUserResp, error)
	DeleteUser(ctx context.Context, userID int64) error
	ImportUsers(ctx context.Context, reader io.Reader) (*dto.ImportUsersResp, error)
}

type AdminQueryService interface {
	ListUsers(ctx context.Context, query *dto.AdminUserQuery) ([]dto.AdminUserResp, int64, int, int, error)
}
