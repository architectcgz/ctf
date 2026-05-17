package contracts

import (
	"context"
	"io"
	"time"
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

type AdminUserListQuery struct {
	Keyword   string
	StudentNo string
	TeacherNo string
	Role      string
	Status    string
	ClassName string
	Page      int
	Size      int
}

type AdminUser struct {
	ID        int64
	Username  string
	Name      *string
	Email     *string
	StudentNo *string
	TeacherNo *string
	ClassName *string
	Status    string
	Roles     []string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type ImportUserError struct {
	Row     int
	Message string
}

type ImportUsersResult struct {
	Created int
	Updated int
	Failed  int
	Errors  []ImportUserError
}

type AdminCommandService interface {
	CreateUser(ctx context.Context, req CreateUserInput) (*AdminUser, error)
	UpdateUser(ctx context.Context, userID int64, req UpdateUserInput) (*AdminUser, error)
	DeleteUser(ctx context.Context, userID int64) error
	ImportUsers(ctx context.Context, reader io.Reader) (*ImportUsersResult, error)
}

type AdminQueryService interface {
	ListUsers(ctx context.Context, query AdminUserListQuery) ([]AdminUser, int64, int, int, error)
}
