package identity

import (
	"context"
	"errors"
	"io"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	jwtpkg "ctf-platform/pkg/jwt"
)

type Authenticator interface {
	authcontracts.TokenService
	ParseAccessToken(token string) (*jwtpkg.Claims, error)
}

var ErrUserNotFound = errors.New("identity user not found")
var ErrUsernameExists = errors.New("identity username already exists")
var ErrEmailExists = errors.New("identity email already exists")
var ErrStudentNoExists = errors.New("identity student no already exists")
var ErrTeacherNoExists = errors.New("identity teacher no already exists")
var ErrRoleNotFound = errors.New("identity role not found")

type UserListFilter struct {
	Keyword   string
	StudentNo string
	TeacherNo string
	Role      string
	Status    string
	ClassName string
	Offset    int
	Limit     int
}

type UserRepository interface {
	List(ctx context.Context, filter UserListFilter) ([]*model.User, int64, error)
	FindByID(ctx context.Context, userID int64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, userID int64) error
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
	UpdateLoginState(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error
	UpdateProfile(ctx context.Context, user *model.User) error
}

type AdminService interface {
	ListUsers(ctx context.Context, query *dto.AdminUserQuery) ([]dto.AdminUserResp, int64, int, int, error)
	CreateUser(ctx context.Context, req *dto.CreateAdminUserReq) (*dto.AdminUserResp, error)
	UpdateUser(ctx context.Context, userID int64, req *dto.UpdateAdminUserReq) (*dto.AdminUserResp, error)
	DeleteUser(ctx context.Context, userID int64) error
	ImportUsers(ctx context.Context, reader io.Reader) (*dto.ImportUsersResp, error)
}

type ProfileService interface {
	GetProfile(ctx context.Context, userID int64) (*dto.AuthUser, error)
	ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordReq) error
}
